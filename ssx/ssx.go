package ssx

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kevinburke/ssh_config"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"

	"github.com/vimiix/ssx/internal/errmsg"
	"github.com/vimiix/ssx/internal/lg"
	"github.com/vimiix/ssx/internal/slice"
	"github.com/vimiix/ssx/internal/tui"
	"github.com/vimiix/ssx/internal/utils"
	"github.com/vimiix/ssx/ssx/bbolt"
	"github.com/vimiix/ssx/ssx/entry"
	"github.com/vimiix/ssx/ssx/env"
)

type CmdOption struct {
	DBFile       string
	EntryID      uint64
	Addr         string
	Tag          string
	IdentityFile string
	JumpServers  string
	Keyword      string
	Command      string
	Timeout      time.Duration
	Port         int
	Unsafe       bool
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

	if err := ssx.initRepo(); err != nil {
		return nil, err
	}
	if err := ssx.loadUserSSHConfig(); err != nil {
		return nil, err
	}
	return ssx, nil
}

func (s *SSX) initRepo() error {
	s.repo = bbolt.NewRepo(s.opt.DBFile)
	lg.Debug("init repo")
	return s.repo.Init()
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
		username, _ := cfg.Get(tags[0], "User")
		keyPath, _ := cfg.Get(tags[0], "IdentityFile")
		e := &entry.Entry{
			Host:    hostname,
			Port:    port,
			User:    username,
			KeyPath: utils.ExpandHomeDir(keyPath),
			Tags:    tags,
			Source:  entry.SourceSSHConfig,
		}
		s.sshEntryMap[e.String()] = e
	}
	return nil
}

func (s *SSX) GetEntry(opt *CmdOption) (*entry.Entry, error) {
	// The order determines the priority
	if opt.Keyword != "" {
		lg.Debug("keyword %q take effect", opt.Keyword)
		return s.searchEntry(opt.Keyword)
	} else if opt.EntryID > 0 {
		lg.Debug("entry id %d take effect", opt.EntryID)
		return s.repo.GetEntry(opt.EntryID)
	} else if opt.Addr != "" {
		lg.Debug("addr %q take effect", opt.Addr)
		return s.parseFuzzyAddr(opt.Addr)
	} else if opt.Tag != "" {
		lg.Debug("tag %q take effect", opt.Tag)
		return s.getEntryByTag(opt.Tag)
	} else {
		lg.Debug("select entry from all")
		return s.selectEntryFromAll()
	}
}

func (s *SSX) Main(ctx context.Context) error {
	e, err := s.GetEntry(s.opt)
	if err != nil {
		return err
	}

	// maybe user want to update keyfile
	if s.opt.IdentityFile != "" {
		e.KeyPath = s.opt.IdentityFile
	}

	client := NewClient(e, s.repo)
	if len(s.opt.Command) > 0 {
		opt := &ExecuteOption{
			Command: s.opt.Command,
			Stdout:  os.Stdout,
			Stderr:  os.Stderr,
			Timeout: s.opt.Timeout,
		}
		return client.Execute(ctx, opt)
	}

	return client.Interact(ctx)
}

func (s *SSX) getAllEntries() ([]*entry.Entry, error) {
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
	return es, nil
}

func (s *SSX) buildNewEntry(host, username, port string) (*entry.Entry, error) {
	if port == "" && s.opt.Port > 0 {
		// Takes effect for new entry only
		port = strconv.Itoa(s.opt.Port)
	}
	safeMode := entry.ModeSafe
	if s.opt.Unsafe || env.IsUnsafeMode() {
		safeMode = entry.ModeUnsafe
	}
	e := &entry.Entry{
		Host:     host,
		User:     username,
		Port:     port,
		KeyPath:  utils.ExpandHomeDir(s.opt.IdentityFile),
		Source:   entry.SourceSSXStore,
		SafeMode: safeMode,
	}
	if s.opt.JumpServers != "" {
		proxy, err := parseProxyChainFromString(s.opt.JumpServers)
		if err != nil {
			return nil, err
		}
		e.Proxy = proxy
	}
	if err := e.Tidy(); err != nil {
		return nil, err
	}
	return e, nil
}

// search by host and tag first, if not found, then connect as a new entry
func (s *SSX) searchEntry(keyword string) (*entry.Entry, error) {
	es, err := s.getAllEntries()
	if err != nil {
		return nil, err
	}
	var candidates []*entry.Entry
	for _, e := range es {
		if utils.ContainsI(e.String(), keyword) ||
			utils.ContainsI(strings.Join(e.Tags, " "), keyword) {
			candidates = append(candidates, e)
		}
	}
	if len(candidates) == 1 {
		lg.Debug("found exist entry: %s", candidates[0].String())
		return candidates[0], nil
	}
	if len(candidates) > 1 {
		return s.selectEntry(candidates)
	}
	lg.Debug("not found by keyword %q, treat it as new entry", keyword)
	match, err := utils.MatchAddress(keyword)
	if err != nil {
		return nil, err
	}
	e, err := s.buildNewEntry(match.Host, match.User, match.Port)
	if err != nil {
		return nil, err
	}
	for _, exist := range es {
		if exist.String() == e.String() {
			lg.Debug("found exist entry: %s", exist.String())
			return exist, nil
		}
	}
	return e, nil
}

func (s *SSX) selectEntryFromAll() (*entry.Entry, error) {
	es, err := s.getAllEntries()
	if err != nil {
		return nil, err
	}
	return s.selectEntry(es)
}

func (s *SSX) parseFuzzyAddr(addr string) (*entry.Entry, error) {
	// [user@]host[:port][/path]
	match, err := utils.MatchAddress(addr)
	if err != nil {
		return nil, err
	}
	username, host, port := match.User, match.Host, match.Port

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
		return s.selectEntry(candidates, "multiple entries found, select one")
	}

	// new entry
	lg.Debug("it is a fresh entry")
	return s.buildNewEntry(host, username, port)
}

func parseProxyChainFromString(s string) (*entry.Proxy, error) {
	if s == "" {
		return nil, nil
	}
	servers := strings.Split(s, ",")
	var (
		root, cursor *entry.Proxy
		err          error
	)
	for idx, server := range servers {
		if idx == 0 {
			root, err = parseProxyFromString(server)
			if err != nil {
				return nil, err
			}
			if root == nil {
				break
			}
			cursor = root
			continue
		}

		p, err := parseProxyFromString(server)
		if err != nil {
			return nil, err
		}
		if p == nil {
			break
		}
		cursor.Proxy = p
		cursor = p
	}
	return root, nil
}

func parseProxyFromString(s string) (*entry.Proxy, error) {
	if s == "" {
		return nil, nil
	}
	match, err := utils.MatchAddress(s)
	if err != nil {
		return nil, err
	}
	p := &entry.Proxy{
		User: match.User,
		Host: match.Host,
		Port: match.Port,
	}
	lg.Debug("parse proxy %s success", s)
	return p, nil
}

func foundTargetByAddr[T comparable](em map[T]*entry.Entry, host, username, port string) (hit *entry.Entry, candidates []*entry.Entry) {
	for _, e := range em {
		if !utils.ContainsI(e.Host, host) {
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
		return nil, errors.Errorf("not found any entry by tag: %q", tag)
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	return s.selectEntry(candidates, "multiple entries found, select one")
}

func foundTargetByTag[T comparable](em map[T]*entry.Entry, tagKeyword string) (candidates []*entry.Entry) {
	for _, e := range em {
		if len(e.Tags) == 0 {
			continue
		}
		for _, tag := range e.Tags {
			if utils.ContainsI(tag, tagKeyword) {
				candidates = append(candidates, e)
				break
			}
		}
	}
	return
}

var templates = &promptui.SelectTemplates{
	Active:   "➤ {{ .User | green }}{{ `@` | green }}{{ .Host | green }}{{if .Tags }} {{ .Tags | faint}}{{ end }}",
	Inactive: "  {{ .User | faint }}{{ `@` | faint }}{{ .Host | faint }}{{if .Tags }} {{ .Tags | faint}}{{ end }}",
}

func (s *SSX) selectEntry(es []*entry.Entry, promptOption ...string) (*entry.Entry, error) {
	if len(es) == 0 {
		return nil, errmsg.ErrNoEntry
	}
	sort.Slice(es, func(i, j int) bool {
		return es[i].ID < es[j].ID
	})
	searcher := func(input string, index int) bool {
		e := es[index]
		content := fmt.Sprintf("%s %s", e.String(), strings.Join(e.Tags, " "))
		return utils.ContainsI(content, input)
	}

	promptStr := "select entry"
	if len(promptOption) > 0 {
		promptStr = promptOption[0]
	}
	prompt := promptui.Select{
		Label:             promptStr,
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
		return errmsg.ErrNoEntry
	}

	if len(repoEntryMap) > 0 {
		var entries []*entry.Entry
		header := []string{"ID", "Address", "Tags"}
		var rows [][]string
		for _, e := range repoEntryMap {
			entries = append(entries, e)
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].ID < entries[j].ID
		})
		for _, e := range entries {
			rows = append(rows,
				[]string{strconv.Itoa(int(e.ID)), e.String(), strings.Join(e.Tags, ",")},
			)
		}

		fmt.Println("Entries (stored in ssx)")
		tui.PrintTable(header, rows)
	}

	if len(s.sshEntryMap) > 0 {
		header := []string{"Address", "Tags"}
		var rows [][]string
		for _, e := range s.sshEntryMap {
			rows = append(rows,
				[]string{e.String(), strings.Join(e.Tags, ",")},
			)
		}
		fmt.Println()
		fmt.Println("Entries (found in ssh config)")
		tui.PrintTable(header, rows)
	}
	return nil
}

func (s *SSX) DeleteEntryByID(ids ...int) error {
	if len(ids) == 0 {
		return nil
	}

	em, err := s.repo.GetAllEntries()
	if err != nil {
		return err
	}
	var deleteMap = map[uint64]struct{}{}
	for _, id := range ids {
		deleteMap[uint64(id)] = struct{}{}
	}
	for _, e := range em {
		if _, exist := deleteMap[e.ID]; exist {
			lg.Info("deleting %d ...", e.ID)
			if deleteErr := s.repo.DeleteEntry(e.ID); deleteErr != nil {
				lg.Error("failed to delete entry %d", e.ID)
				return deleteErr
			}
			lg.Info("entry %d deleted", e.ID)
		}
	}
	return nil
}

func (s *SSX) DeleteTagByID(id int, tags ...string) error {
	if len(tags) == 0 {
		return nil
	}
	em, err := s.repo.GetAllEntries()
	if err != nil {
		return err
	}
	lg.Debug("deleting tags %s for id %d", tags, id)
	for _, e := range em {
		if int(e.ID) != id {
			continue
		}
		e.Tags = slice.Delete(e.Tags, tags...)
		if err = s.repo.TouchEntry(e); err != nil {
			return err
		}
		lg.Info("tags %s deleted", tags)
		return nil
	}
	return errmsg.ErrEntryNotExist
}

func (s *SSX) AppendTagByID(id int, tags ...string) error {
	if len(tags) == 0 {
		return nil
	}
	var reserved []string
	for _, tag := range tags {
		if isReservedWord(tag) {
			reserved = append(reserved, tag)
		}
	}
	if len(reserved) > 0 {
		return fmt.Errorf("can not contain reserved words: %s\nsee also %s",
			reserved, "https://github.com/vimiix/ssx/issues/14")
	}

	em, err := s.repo.GetAllEntries()
	if err != nil {
		return err
	}
	lg.Debug("adding tags %s for id %d", tags, id)
	for _, e := range em {
		if int(e.ID) != id {
			continue
		}
		e.Tags = slice.Union(e.Tags, tags)
		if err = s.repo.TouchEntry(e); err != nil {
			return err
		}
		lg.Info("tags %s added", tags)
		return nil
	}
	return errmsg.ErrEntryNotExist
}
