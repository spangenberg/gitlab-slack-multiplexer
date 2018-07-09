# GitLab Slack Multiplexer

## Getting Started

### Slack command configuration

Follow the standard [GitLab documentation](https://docs.gitlab.com/ce/user/project/integrations/slack_slash_commands.html) and instead of passing the URL GitLab suggests use the URL of where the multiplexer is running with the path `/slack/command`.

Example: `https://gitlab-slack-multiplexer.example.com/slack/command`

### Running

Pass the URL of the GitLab installation as parameter or environment variable to the process.

* --gitlab-url=https://gitlab.example.com
* GITLAB_URL=https://gitlab.example.com

#### Docker

`docker run --rm -d -p 8080:8080 -e GITLAB_URL=https://gitlab.example.com spangenberg/gitlab-slack-multiplexer`
