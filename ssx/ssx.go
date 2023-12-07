package ssx

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/kevinburke/ssh_config"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"

	"github.com/vimiix/ssx/internal/cleaner"
	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/internal/tui"
	"github.com/vimiix/ssx/internal/utils"
	"github.com/vimiix/ssx/ssx/bbolt"
	"github.com/vimiix/ssx/ssx/entry"
	"github.com/vimiix/ssx/ssx/env"
)

type CmdOption struct {
	DBFile string
	Addr   string
	Tag    string
}

// Tidy complete unset fields with default values
func (o *CmdOption) Tidy() error {
	if len(o.DBFile) <= 0 {
		envDBPath := os.Getenv(env.SSXDBPath)
		if len(envDBPath) > 0 {
			lg.Debug("env %q taking effect", envDBPath)
			o.DBFile = envDBPath
		} else {
			u, err := user.Current()
			if err != nil {
				return err
			}
			o.DBFile = path.Join(u.HomeDir, ".ssx.db")
		}
	}
	return nil
}

type SSX struct {
	opt         *CmdOption
	repo        Repo
	sshEntryMap map[string]*entry.Entry
}

func NewSSX(opt *CmdOption) (*SSX, error) {
	if err := opt.Tidy(); err != nil {
		return nil, err
	}
	ssx := &SSX{opt: opt}

	if err := ssx.openRepo(); err != nil {
		return nil, err
	}
	if err := ssx.loadUserSSHConfig(); err != nil {
		return nil, err
	}
	return ssx, nil
}

func (s *SSX) openRepo() error {
	s.repo = bbolt.NewRepo(s.opt.DBFile)
	lg.Debug("open repo")
	if err := s.repo.Open(); err != nil {
		return err
	}
	cleaner.RegisterCallback(func() {
		lg.Debug("close repo")
		_ = s.repo.Close()
	})
	return nil
}

func (s *SSX) loadUserSSHConfig() error {
	if os.Getenv(env.SSXImportSSHConfig) == "" {
		lg.Debug("not found env %q, skip load user ssh config", env.SSXImportSSHConfig)
		return nil
	}
	s.sshEntryMap = map[string]*entry.Entry{}
	u, err := user.Current()
	if err != nil {
		return err
	}
	sshConfigFile := filepath.Join(u.HomeDir, ".ssh/config")
	if !utils.FileExists(sshConfigFile) {
		lg.Debug("user ssh config not exist")
		return nil
	}
	lg.Debug("parsing user ssh config: %s", sshConfigFile)
	f, err := os.Open(sshConfigFile)
	if err != nil {
		return err
	}
	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return err
	}
	for _, host := range cfg.Hosts {
		var tags []string
		for _, p := range host.Patterns {
			tags = append(tags, p.String())
		}
		hostname, _ := cfg.Get(tags[0], "HostName")
		if hostname == "" {
			continue
		}
		lg.Debug("processing: %q", hostname)
		port, _ := cfg.Get(tags[0], "Port")
		if port != "" {
			_, err = strconv.Atoi(port)
			if err != nil {
				return errors.Wrapf(err, "invalid port value %q of %q", port, hostname)
			}
		} else {
			port = "22"
		}
		user, _ := cfg.Get(tags[0], "User")
		keyPath, _ := cfg.Get(tags[0], "IdentityFile")
		e := &entry.Entry{
			Host:    hostname,
			Port:    port,
			User:    user,
			KeyPath: utils.ExpandHomeDir(keyPath),
			Tags:    tags,
		}
		s.sshEntryMap[e.UniqueKey()] = e
	}
	return nil
}

func (s *SSX) Main(ctx context.Context) error {
	var (
		e   *entry.Entry
		err error
	)
	if s.opt.Addr != "" {
		e, err = s.parseFuzzyAddr(s.opt.Addr)
	} else if s.opt.Tag != "" {
		e, err = s.getEntryByTag(s.opt.Tag)
	} else {
		e, err = s.SelectEntryFromAll()
	}
	if err != nil {
		return err
	}

	return s.Run(e)
}

func (s *SSX) Run(e *entry.Entry) error {
	cli := NewClient(e)
	return cli.Run()
}

func (s *SSX) SelectEntryFromAll() (*entry.Entry, error) {
	var es []*entry.Entry
	em, err := s.repo.GetAllEntries()
	if err != nil {
		return nil, err
	}
	for _, e := range em {
		es = append(es, e)
	}
	if len(s.sshEntryMap) > 0 {
		for _, e := range s.sshEntryMap {
			es = append(es, e)
		}
	}
	return s.selectEntry(es)
}

var addrRegex = regexp.MustCompile(`^(?:(?P<user>\w+)@)?(?P<host>[\w.-]+)(?:\:(?P<port>\d+))?(?:\/(?P<path>[\w\/.-]+))?$`)

func (s *SSX) parseFuzzyAddr(addr string) (*entry.Entry, error) {
	// [user@]host[:port][/path]
	matches := addrRegex.FindStringSubmatch(addr)
	if len(matches) == 0 {
		return nil, errors.Errorf("invalid address: %s", addr)
	}
	username, host, port := matches[1], matches[2], matches[3]

	em, err := s.repo.GetAllEntries()
	if err != nil {
		return nil, err
	}
	hit, candidates := foundTargetByAddr(em, host, username, port)
	if hit != nil {
		return hit, nil
	}
	if len(s.sshEntryMap) > 0 {
		hit, sshCandidates := foundTargetByAddr(s.sshEntryMap, host, username, port)
		if hit != nil {
			return hit, nil
		}
		candidates = append(candidates, sshCandidates...)
	}
	if len(candidates) > 0 {
		if len(candidates) == 1 {
			return candidates[0], nil
		}
		return s.selectEntry(candidates)
	}

	// new entry
	e := &entry.Entry{
		Host: host,
		User: username,
		Port: port,
	}
	if err = e.Tidy(); err != nil {
		return nil, err
	}
	return e, nil
}

func foundTargetByAddr(em map[string]*entry.Entry, host, username, port string) (hit *entry.Entry, candidates []*entry.Entry) {
	for _, e := range em {
		if e.Host != host {
			continue
		}
		if username != "" && e.User == username {
			// almost found it
			if port != "" && e.Port != port {
				e.Port = port
			}
			hit = e
			return
		}
		candidates = append(candidates, e)
	}
	return
}

func (s *SSX) getEntryByTag(tag string) (*entry.Entry, error) {
	candidates := foundTargetByTag(s.sshEntryMap, tag)
	em, err := s.repo.GetAllEntries()
	if err != nil {
		return nil, err
	}
	candidates = append(candidates, foundTargetByTag(em, tag)...)
	if len(candidates) == 0 {
		return nil, errors.Errorf("not found any server by tag: %q", tag)
	}
	if len(candidates) == 1 {
		return candidates[1], nil
	}
	return s.selectEntry(candidates)
}

func foundTargetByTag(em map[string]*entry.Entry, tagKeyword string) (candidates []*entry.Entry) {
	for _, e := range em {
		if len(e.Tags) == 0 {
			continue
		}
		for _, tag := range e.Tags {
			if strings.Contains(tag, tagKeyword) {
				candidates = append(candidates, e)
				break
			}
		}
	}
	return
}

var templates = &promptui.SelectTemplates{
	Active:   "➤ {{ .User`@`.Host | green }}{{if .Tags }}({{ .Tags | faint}}){{ end }}",
	Inactive: "  {{ .User`@`.Host | faint }}{{if .Tags }}({{ .Tags | faint}}){{ end }}",
}

func (s *SSX) selectEntry(es []*entry.Entry) (*entry.Entry, error) {
	searcher := func(input string, index int) bool {
		e := es[index]
		content := fmt.Sprintf("%s %s", e.String(), strings.Join(e.Tags, " "))
		return strings.Contains(content, input)
	}
	prompt := promptui.Select{
		Label:             "select server:",
		Items:             es,
		Size:              20,
		HideSelected:      true,
		Templates:         templates,
		Searcher:          searcher,
		StartInSearchMode: true,
	}
	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	return es[idx], nil
}

func (s *SSX) ListEntries() error {
	repoEntryMap, err := s.repo.GetAllEntries()
	if err != nil {
		return err
	}
	if len(repoEntryMap) == 0 && len(s.sshEntryMap) == 0 {
		fmt.Println("Nothing found")
		return nil
	}

	if len(repoEntryMap) > 0 {
		header := []string{"ID", "Address", "Tags"}
		var rows [][]string
		for _, entry := range repoEntryMap {
			rows = append(rows,
				[]string{strconv.Itoa(int(entry.ID)), entry.String(), strings.Join(entry.Tags, ",")},
			)
		}
		fmt.Println("Entries (stored by ssx)")
		tui.PrintTable(header, rows)
	}

	if len(s.sshEntryMap) > 0 {
		header := []string{"Address", "Tags"}
		var rows [][]string
		for _, entry := range s.sshEntryMap {
			rows = append(rows,
				[]string{entry.String(), strings.Join(entry.Tags, ",")},
			)
		}
		fmt.Println("Entries (found in ssh config)")
		tui.PrintTable(header, rows)
	}
	return nil
}

func (s *SSX) DeleteEntryByID(ctx context.Context, ids ...int) error {
	return nil
}
