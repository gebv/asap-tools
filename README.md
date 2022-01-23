# asap-tools

[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/George.Bach)
[![GitHub Release](https://img.shields.io/github/release/gebv/asap-tools)](https://github.com/gebv/asap-tools/releases/latest)

The `asap-tools` it is collection of tools to simplify daily monotonous monotonous cases.

In the arsenal today:
- [go to](https://github.com/gebv/asap-tools#sync-clickup) syncing tasks (mirror tasks) between ClickUp teams
- WIP saving conversations from Slack
- TODO create github action for very quick starts for the periodic runs of asap-tools
- TODO syncing tasks between ClickUp and Notion
- TODO cross likes between Spotify and Last.fm
- TODO backup conversations from Telegram direct chat

## Sync ClickUp

Features
- create mirror-task and sync (TODO more details)
- Firestore (database from Google Firebase) is used as permanent storage

![asap-tools sync with clickup ](.github/clickup-preview.gif)

[Guide for quick start](clickup/README.md)

TODO:
- (draft) magic-action comments and syncing comments
- (draft) sync with another task tracker (GitHub, ...)
- (draft) hook from changed task - send to another task tracker (GitHub, ...)
- (draft) hook from changed task - send to messenger (telegram, ...)
- (draft) support for custom fields (really necessary?)

You can help (contact me via github issues)
- add new API methods or expand models (add missing fields)
- implement webhooks
- offer a new features to the arsenal of the sync ClickUp tasks
- bug reports are welcome
- writing e2e tests (manual testing is tired)

---

## For The Developer

How to develop custom storage models read more [here](storage/README.md)
An example of use [here](clickup/model.go)
