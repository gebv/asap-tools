package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/gebv/asap-tools/clickup"
	clickupAPI "github.com/gebv/asap-tools/clickup/api"
	"github.com/gebv/asap-tools/logger"
	"github.com/gebv/asap-tools/storage"
	"github.com/gebv/asap-tools/version"

	"cloud.google.com/go/firestore"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
)

var (
	onlyShowVersionF = flag.Bool("v", false, "Only shows version.")
	onlyShowHelpF    = flag.Bool("help", false, "Only shows help.")

	clickupCommands            = flag.NewFlagSet("clickup", flag.ExitOnError)
	clickupDebugExampleSpecF   = clickupCommands.Bool("debug-example-spec", false, "Shows an example of a spec of sync in yaml format.")
	clickupRecentActivitySyncF = clickupCommands.Bool("recent-activity-sync", false, "Regular procedure for loading changed tasks from ClickUp API and processing.")
	clickupDBSyncF             = clickupCommands.Bool("db-sync", false, "Foce loads all tasks from the database and processing. To use if the spec of sync file has been changed. (NOTE: Changed tasks are not loaded from the ClickUp api). ")
)

func printAllFlagUsage() {
	fmt.Println("Available commands and flags:")
	for _, item := range []interface {
		Name() string
		PrintDefaults()
	}{clickupCommands} {
		fmt.Printf("- command %q with flags:\n", item.Name())
		item.PrintDefaults()
	}
	fmt.Println("And global flags:")
	flag.PrintDefaults()
}

func unknownCommandAndExist() {
	fmt.Println("Unknown command")
	printAllFlagUsage()
	log.Fatalln()
}

var (
	Ctx context.Context
	Cfg *Config
)

func showVersion() {
	fmt.Printf("AsapTools cli version %s (%s_%s) git commit %s\n", version.Version, version.Goos, version.Goarch, version.GitCommit)
	fmt.Println("Build Date:", version.BuildDate)
	fmt.Println("Start Date:", time.Now().UTC().Format(time.RFC3339))
	fmt.Println()
	fmt.Println("https://github.com/gebv/asap-tools")
}

func showHelp() {
	fmt.Println()
	envconfig.Usage(configNamespace, &Config{})
	fmt.Println()
	printAllFlagUsage()
}

func main() {
	flag.Parse()

	showVersion()
	if *onlyShowHelpF {
		showHelp()
	}
	if *onlyShowVersionF || *onlyShowHelpF {
		return
	}

	cfg := &Config{}
	ParseOrPanic(cfg)

	logger.Setup(cfg.LoggerLevel, cfg.DevelopLogger)

	defer func() {
		log.Println("Bye-Bye")
	}()

	if len(os.Args) < 2 {
		unknownCommandAndExist()
	}

	switch os.Args[1] {
	case clickupCommands.Name():
		clickupCommands.Parse(os.Args[2:])
	default:
		unknownCommandAndExist()
	}

	// TODO : handle term
	Ctx = context.Background()

	if clickupCommands.Parsed() {
		handleClickupCommands()
	}
}

func clickupShowDemoSpec() {
	spec := &clickup.SyncPreferences{
		MirrorTaskRules: []clickup.MirrorTaskSpecification{
			{
				Name: "<NameRule>",
				CondAdd: &clickup.SyncRule_CondOfAdd{
					IfInLists: []string{
						"https://app.clickup.com/<TeamID>/v/li/<ListID>",
					},
				},
				CondTrackChanges: &clickup.SyncRule_CondTrackChanges{
					IfInLists: []string{
						"https://app.clickup.com/<TeamID>/v/li/<ListID>",
					},
				},
				SpecAdd: &clickup.SyncRule_SpecOfAdd{
					AddToList: "https://app.clickup.com/<TeamID>/v/li/<ListID>",
				},
			},
		},
	}
	specBytes, _ := yaml.Marshal(spec)
	fmt.Println("Example of spec yaml file for sync ClickUp tasks.")
	fmt.Println()
	fmt.Println(string(specBytes))
	fmt.Println()
}

func handleClickupCommands() {

	if *clickupDebugExampleSpecF {
		clickupShowDemoSpec()
		return
	}

	firestoreOpts := []option.ClientOption{}
	if Cfg.Firestore.CredsInlineJSON != "" {
		firestoreOpts = append(firestoreOpts, option.WithCredentialsJSON([]byte(Cfg.Firestore.CredsInlineJSON)))
	}

	client, err := firestore.NewClient(Ctx, Cfg.Firestore.ProjectID, firestoreOpts...)
	if err != nil {
		zap.L().Error("Failed setup firestore client", zap.Error(err))
		return
	}

	spec := &clickup.SyncPreferences{}
	{
		specBytes, err := ioutil.ReadFile(Cfg.Clickup.FileSpecSync)
		if err != nil {
			zap.L().Fatal("Failed read spec of sync file", zap.Error(err), zap.String("file_path", Cfg.Clickup.FileSpecSync))
		}

		if err := yaml.Unmarshal(specBytes, spec); err != nil {
			zap.L().Fatal("Failed decode spec of sync from yaml", zap.String("file_path", Cfg.Clickup.FileSpecSync),
				zap.Error(err), zap.String("file_raw", string(specBytes)))
		}
	}

	storage := storage.NewStorage(client)
	clickupStorage := clickup.NewStorage(storage)
	api := clickupAPI.NewAPI(Cfg.Clickup.ApiToken)
	manage := clickup.NewChangeManager(api, clickupStorage)

	if *clickupRecentActivitySyncF {
		teamIDs := spec.AllUsedTeamIDs()

		zap.L().Info("[PROCESS_EXISTS_TASKS] processing of existing tasks in the database for teams from spec sync", zap.Any("team_ids", teamIDs))
		for _, teamID := range spec.AllUsedTeamIDs() {
			list := clickupStorage.AllTeamTasks(Ctx, teamID)
			for idx := range list {
				task := list[idx]

				// forced processing
				manage.Sync(Ctx, spec, task, task, true)
			}
		}
	}

	if *clickupRecentActivitySyncF {
		teamIDs := spec.AllUsedTeamIDs()
		zap.L().Info("processing of the last changed tasks (from ClickUp API) for teams from spec sync", zap.Any("team_ids", teamIDs))
		for _, teamID := range spec.AllUsedTeamIDs() {
			manage.ApplyChangesInTeam(Ctx, spec, teamID)
		}
	}

}

type Config struct {
	DevelopLogger bool   `envconfig:"LOG_DEV" default:"false"`
	LoggerLevel   string `envconfig:"LOG_LEVEL" default:"WARN" desc:"Logging level (availabel DEBUG, INFO, WARN, ERROR)"`

	Firestore *FirestoreSettings `envconfig:"FIRESTORE"`
	Clickup   *ClickupConfig     `envconfig:"CLICKUP"`
}

type ClickupConfig struct {
	ApiToken      string `envconfig:"API_TOKEN" desc:"Token from ClickUp API (follow link https://app.clickup.com/settings/apps)"`
	WebhookSecret string `envconfig:"WEBHOOK_SECRET" ignored:"true" desc:"Webhook secret from ClickUp API (follow link https://clickup20.docs.apiary.io/#reference/0/webhooks secion 'Signature')."` // WIP
	FileSpecSync  string `envconfig:"FILE_SPEC_SYNC"`
}

type FirestoreSettings struct {
	CredsInlineJSON string `envconfig:"PRIVATE_KEY_INLINE_JSON" desc:"Inline json file with Google Cloud service account private key."`
	ProjectID       string `envconfig:"PROJECT_ID" desc:"Google Cloud project ID"`
}

const configNamespace = "ASAPTOOLS"

func ParseOrPanic(cfg *Config) {
	err := envconfig.Process(configNamespace, cfg)
	if err != nil {
		// TODO: custom template
		envconfig.Usage(configNamespace, cfg)
		fmt.Println()
		fmt.Println("Failed process parse config from env:", err)
		os.Exit(1)
	}
}
