![asap-tools sync with clickup ](../.github/clickup-schema.png)

# quick start

Install `asap-tools` via brew

```bash
brew install gebv/tap/asap-tools
# asap-tools-cli -v
# AsapTools cli version 0.0.3 (darwin_amd64) git commit 40401abab24202e19852464d953e16e99e77125e
# Build Date: 2022-01-23T05:27:34Z
# Start Date: 2022-01-23T05:32:02Z
# https://github.com/gebv/asap-tools
```

or download binary files from the [latest release](https://github.com/gebv/asap-tools/releases/latest)

```bash
curl -L https://github.com/gebv/asap-tools/releases/latest/download/asap-tools_Linux_x86_64.zip > ./file.zip && \
  unzip ./file.zip
chmod +x ./asap-tools-cli
sudo mv ./asap-tools-cli /usr/local/bin/asap-tools-cli
asap-tools-cli -v
# AsapTools cli version 0.0.6 (linux_amd64) git commit 598374dedea93708f9b5fc49178b0bd7bea32d6f
# Build Date: 2022-01-23T07:58:37Z
# Start Date: 2022-01-23T08:18:30Z

# https://github.com/gebv/asap-tools
```

Configuring the spec file (command for example yaml file `asap-tools-cli clickup -debug-example-spec`).

```yaml
mirror_task_rules:
- name: <NameRule>
  # conds for adding new tasks from the team with the original tasks
  cond_add:
    # list folders of interest
    if_in_folders: []
    # list lists of interest
    if_in_lists:
    - https://app.clickup.com/<TeamID>/v/li/<ListID>
    # only the specified task statuses or all tasks
    eq_any_task_status_names: []
    # if assigned task to member
    if_assigned_to_member_email: ""
  # conds for track changes from the team with the original tasks
  cond_track_changes:
    if_in_folders: []
    if_in_lists:
    - https://app.clickup.com/<TeamID>/v/li/<ListID>
    eq_any_task_status_names: []
    if_assigned_to_member_email: ""
  # spec for adding new mirror tasks
  spec_add:
    # mirror tasks are added to the list
    add_to_list: https://app.clickup.com/<TeamID>/v/li/<ListID>
    set_status_name: ""
    assign_to_member_email: ""
# status association
global_mirror_task_statuses:
  # status "done" in mirror task says
  done:
    # not sync estimate and due date
    sync_estimate: false
    # orig task status will be set to "ready"
    orig_task_status: ready
  open:
    sync_estimate: true
    orig_task_status: in progress
  wip:
    sync_estimate: true
    orig_task_status: in progress
```

Set the necessary envs (current on 2021-01-23, show actual envs and commands via command `asap-tools-cli -help`)

```csv
KEY                                            TYPE             DEFAULT    REQUIRED    DESCRIPTION
ASAPTOOLS_LOG_DEV                              True or False    false
ASAPTOOLS_LOG_LEVEL                            String           WARN                   Logging level (availabel DEBUG, INFO, WARN, ERROR)
ASAPTOOLS_FIRESTORE_PRIVATE_KEY_INLINE_JSON    String                                  Inline json file with Google Cloud service account private key.
ASAPTOOLS_FIRESTORE_PROJECT_ID                 String                                  Google Cloud project ID
ASAPTOOLS_CLICKUP_API_TOKEN                    String                                  Token from ClickUp API (follow link https://app.clickup.com/settings/apps)
ASAPTOOLS_CLICKUP_FILE_SPEC_SYNC               String
```

Run a command to retrieve changed tasks and processing them.

```bash
asap-tools-cli clickup -recent-activity-sync
```

After each spec file change, run the command (to upgrade and processing to existing tasks)

```bash
asap-tools-cli clickup -db-sync
```
