# Why SSX

As a backend developer, I frequently work with numerous servers (conservatively speaking, I interact with 50+ servers daily). SSH is an indispensable development tool. However, having to enter passwords every time I login - especially for servers where I can't set up key-based authentication because they're just for temporary troubleshooting - is somewhat unacceptable for a programmer who believes everything should be automated (if you use GUI tools, you can ignore this).

Before developing SSX, I thought about what I needed: an SSH client that doesn't need many complex features, just these few requirements for daily use:

- Similar usage habits to standard ssh
- Only ask for password on first login, no password needed for subsequent logins
- Ability to tag servers freely, so I can login via IP or tag

So in my spare time, I designed and developed SSX - a lightweight SSH client with memory. It perfectly implements all the features I needed and has been happily integrated into my daily development workflow, greatly improving my productivity.
