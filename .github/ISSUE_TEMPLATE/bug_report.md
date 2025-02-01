---
name: Bug report
about: Create a report to help us improve go-hass-agent
title: "[BUG]"
labels: bug
assignees: ''

---

**Go Hass Agent Version**
Retrieve with `go-hass-agent version`.

**Describe the bug**
A clear and concise description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior:

1. Do X.
2. Do Y.
3. …
4. See error.

**Expected behavior**
A clear and concise description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Logs**
A log will be very helpful to look into this bug report. To get the log:

1. Run `go-hass-agent` from a terminal or command-line with the `--log-level debug` flag set:

  ```shell
  # for package/binary installs:
  go-hass-agent --log-level=debug run
  # for containers:
  # (adjust options as appropriate)
  podman run --hostname _HOSTNAME_ --name go-hass-agent \
    --network host \
    --volume go-hass-agent:/home/go-hass-agent:U \
    # add any additional container options here.
    --device /dev/video0:/dev/video0
    ghcr.io/joshuar/go-hass-agent:VERSION \
    # add any Go Hass Agent options here.
  ```

2. Try to reproduce the issue.
3. After you have reproduced the issue, please (compress and) attach the
   `go-hass-agent.log` file `~/.config/go-hass-agent/go-hass-agent.log`.

*(While I have made efforts to not log sensitive information, please check the log before uploading to GitHub and remove any information you do not want to share).*

**Desktop (please complete the following information):**

- OS: [e.g., Linux]
- Distribution [for Linux, e.g., Fedora]
- Version [e.g., 11]

**Additional context**
Add any other context about the problem here.
