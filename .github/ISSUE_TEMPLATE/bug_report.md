---
name: Bug report
about: Create a report to help us improve
title: "[BUG]"
labels: ''
assignees: joshuar

---

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '…'
2. Click on '…'
3. Scroll down to '…'
4. See error

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Logs**
A log will be very helpful to look into this bug report. To get the log:

1. Run `go-hass-agent` from a terminal or command-line with the `--debug` flag:
```shell
go-hass-agent --debug
```
2. Try to reproduce the problem.
3. After you have reproduced the problem, please (compress and) attach the `go-hass-agent.log` file found in the following location:
  - On Linux, in `~/.config/fyne/com.github.joshuar.go-hass-agent/go-hass-app.log`.

*(While I have made efforts to not log sensitive information, please check the log before uploading to GitHub and remove any information you do not want to share).*

**Desktop (please complete the following information):**
 - OS: [e.g., iOS]
 - Browser [e.g., chrome, safari]
 - Version [e.g., 22]

**Additional context**
Add any other context about the problem here.
