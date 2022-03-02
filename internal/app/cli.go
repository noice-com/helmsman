package app

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/imdario/mergo"
	"github.com/joho/godotenv"
)

const (
	banner = "\n" +
		" _          _ \n" +
		"| |        | | \n" +
		"| |__   ___| |_ __ ___  ___ _ __ ___   __ _ _ __\n" +
		"| '_ \\ / _ \\ | '_ ` _ \\/ __| '_ ` _ \\ / _` | '_ \\ \n" +
		"| | | |  __/ | | | | | \\__ \\ | | | | | (_| | | | | \n" +
		"|_| |_|\\___|_|_| |_| |_|___/_| |_| |_|\\__,_|_| |_|"
	slogan = "A Helm-Charts-as-Code tool.\n\n"
)

// Allow parsing of multiple string command line options into an array of strings
type stringArray []string

type fileOptionArray []fileOption

type fileOption struct {
	name     string
	priority int
}

func (f *fileOptionArray) String() string {
	var a []string
	for _, v := range *f {
		a = append(a, v.name)
	}
	return strings.Join(a, " ")
}

func (f *fileOptionArray) Set(value string) error {
	var fo fileOption

	fo.name = value
	*f = append(*f, fo)
	return nil
}

func (f fileOptionArray) sort() {
	log.Verbose("Sorting files listed in the -spec file based on their priorities... ")

	sort.SliceStable(f, func(i, j int) bool {
		return (f)[i].priority < (f)[j].priority
	})
}

func (i *stringArray) String() string {
	return strings.Join(*i, " ")
}

func (i *stringArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type cli struct {
	debug                 bool
	files                 fileOptionArray
	spec                  string
	envFiles              stringArray
	target                stringArray
	group                 stringArray
	kubeconfig            string
	apply                 bool
	destroy               bool
	dryRun                bool
	verbose               bool
	noBanner              bool
	noColors              bool
	noFancy               bool
	noNs                  bool
	nsOverride            string
	contextOverride       string
	skipValidation        bool
	keepUntrackedReleases bool
	showDiff              bool
	diffContext           int
	noEnvSubst            bool
	substEnvValues        bool
	noSSMSubst            bool
	substSSMValues        bool
	updateDeps            bool
	forceUpgrades         bool
	renameReplace         bool
	version               bool
	noCleanup             bool
	migrateContext        bool
	parallel              int
	alwaysUpgrade         bool
	noUpdate              bool
	kubectlDiff           bool
	checkForChartUpdates  bool
}

func printUsage() {
	fmt.Print(banner)
	fmt.Printf("Helmsman version: " + appVersion + "\n")
	fmt.Printf("Helmsman is a Helm Charts as Code tool which allows you to automate the deployment/management of your Helm charts.")
	fmt.Printf("")
	fmt.Printf("Usage: helmsman [options]\n")
	flag.PrintDefaults()
}

// Cli parses cmd flags, validates them and performs some initializations
func (c *cli) parse() {
	// parsing command line flags
	flag.Var(&c.files, "f", "desired state file name(s), may be supplied more than once to merge state files")
	flag.Var(&c.envFiles, "e", "file(s) to load environment variables from (default .env), may be supplied more than once")
	flag.Var(&c.target, "target", "limit execution to specific app.")
	flag.Var(&c.group, "group", "limit execution to specific group of apps.")
	flag.IntVar(&c.diffContext, "diff-context", -1, "number of lines of context to show around changes in helm diff output")
	flag.IntVar(&c.parallel, "p", 1, "max number of concurrent helm releases to run")
	flag.StringVar(&c.spec, "spec", "", "specification file name, contains locations of desired state files to be merged")
	flag.StringVar(&c.kubeconfig, "kubeconfig", "", "path to the kubeconfig file to use for CLI requests")
	flag.StringVar(&c.nsOverride, "ns-override", "", "override defined namespaces with this one")
	flag.StringVar(&c.contextOverride, "context-override", "", "override releases context defined in release state with this one")
	flag.BoolVar(&c.apply, "apply", false, "apply the plan directly")
	flag.BoolVar(&c.dryRun, "dry-run", false, "apply the dry-run option for helm commands.")
	flag.BoolVar(&c.destroy, "destroy", false, "delete all deployed releases.")
	flag.BoolVar(&c.version, "v", false, "show the version")
	flag.BoolVar(&c.debug, "debug", false, "show the debug execution logs and actual helm/kubectl commands. This can log secrets and should only be used for debugging purposes.")
	flag.BoolVar(&c.verbose, "verbose", false, "show verbose execution logs.")
	flag.BoolVar(&c.noBanner, "no-banner", false, "don't show the banner")
	flag.BoolVar(&c.noColors, "no-color", false, "don't use colors")
	flag.BoolVar(&c.noFancy, "no-fancy", false, "don't display the banner and don't use colors")
	flag.BoolVar(&c.noNs, "no-ns", false, "don't create namespaces")
	flag.BoolVar(&c.skipValidation, "skip-validation", false, "skip desired state validation")
	flag.BoolVar(&c.keepUntrackedReleases, "keep-untracked-releases", false, "keep releases that are managed by Helmsman from the used DSFs in the command, and are no longer tracked in your desired state.")
	flag.BoolVar(&c.showDiff, "show-diff", false, "show helm diff results. Can expose sensitive information.")
	flag.BoolVar(&c.noEnvSubst, "no-env-subst", false, "turn off environment substitution globally")
	flag.BoolVar(&c.substEnvValues, "subst-env-values", false, "turn on environment substitution in values files.")
	flag.BoolVar(&c.noSSMSubst, "no-ssm-subst", false, "turn off SSM parameter substitution globally")
	flag.BoolVar(&c.substSSMValues, "subst-ssm-values", false, "turn on SSM parameter substitution in values files.")
	flag.BoolVar(&c.updateDeps, "update-deps", false, "run 'helm dep up' for local charts")
	flag.BoolVar(&c.forceUpgrades, "force-upgrades", false, "use --force when upgrading helm releases. May cause resources to be recreated.")
	flag.BoolVar(&c.renameReplace, "replace-on-rename", false, "Uninstall the existing release when a chart with a different name is used.")
	flag.BoolVar(&c.noCleanup, "no-cleanup", false, "keeps any credentials files that has been downloaded on the host where helmsman runs.")
	flag.BoolVar(&c.migrateContext, "migrate-context", false, "updates the context name for all apps defined in the DSF and applies Helmsman labels. Using this flag is required if you want to change context name after it has been set.")
	flag.BoolVar(&c.alwaysUpgrade, "always-upgrade", false, "upgrade release even if no changes are found")
	flag.BoolVar(&c.noUpdate, "no-update", false, "skip updating helm repos")
	flag.BoolVar(&c.kubectlDiff, "kubectl-diff", false, "use kubectl diff instead of helm diff. Defalts to false if the helm diff plugin is installed.")
	flag.BoolVar(&c.checkForChartUpdates, "check-for-chart-updates", false, "compares the chart versions in the state file to the latest versions in the chart repositories and shows available updates")
	flag.Usage = printUsage
	flag.Parse()

	if c.version {
		fmt.Println("Helmsman version: " + appVersion)
		os.Exit(0)
	}

	if c.noFancy {
		c.noColors = true
		c.noBanner = true
	}
	verbose := c.verbose || c.debug
	initLogs(verbose, c.noColors)

	if !c.noBanner {
		fmt.Printf("%s version: %s\n%s", banner, appVersion, slogan)
	}

	if c.dryRun && c.apply {
		log.Fatal("--apply and --dry-run can't be used together.")
	}

	if c.destroy && c.apply {
		log.Fatal("--destroy and --apply can't be used together.")
	}

	if len(c.target) > 0 && len(c.group) > 0 {
		log.Fatal("--target and --group can't be used together.")
	}

	if len(flags.files) > 0 && len(flags.spec) > 0 {
		log.Fatal("-f and -spec can't be used together.")
	}

	if c.parallel < 1 {
		c.parallel = 1
	}

	log.Verbose("Helm client version: " + strings.TrimSpace(getHelmVersion()))
	if checkHelmVersion("<3.0.0") {
		log.Fatal("this version of Helmsman does not work with helm releases older than 3.0.0")
	}

	kubectlVersion := getKubectlVersion()
	log.Verbose("kubectl client version: " + kubectlVersion)

	if len(c.files) == 0 && len(c.spec) == 0 {
		log.Info("No desired state files provided.")
		os.Exit(0)
	}

	if c.kubeconfig != "" {
		os.Setenv("KUBECONFIG", c.kubeconfig)
	}

	if !ToolExists(kubectlBin) {
		log.Fatal("kubectl is not installed/configured correctly. Aborting!")
	}

	if !ToolExists(helmBin) {
		log.Fatal("" + helmBin + " is not installed/configured correctly. Aborting!")
	}

	if !c.kubectlDiff && !helmPluginExists("diff") {
		c.kubectlDiff = true
		log.Warning("helm diff not found, using kubectl diff")
	}

	if !c.noEnvSubst {
		log.Verbose("Substitution of env variables enabled")
		if c.substEnvValues {
			log.Verbose("Substitution of env variables in values enabled")
		}
	}
	if !c.noSSMSubst {
		log.Verbose("Substitution of SSM variables enabled")
		if c.substSSMValues {
			log.Verbose("Substitution of SSM variables in values enabled")
		}
	}
}

// readState gets the desired state from files
func (c *cli) readState(s *state) error {
	// read the env file
	if len(c.envFiles) == 0 {
		if _, err := os.Stat(".env"); err == nil {
			err = godotenv.Load()
			if err != nil {
				return fmt.Errorf("error loading .env file: %w", err)
			}
		}
	}

	for _, e := range c.envFiles {
		err := godotenv.Load(e)
		if err != nil {
			return fmt.Errorf("error loading %s env file: %w", e, err)
		}
	}

	// wipe & create a temporary directory
	os.RemoveAll(tempFilesDir)
	_ = os.MkdirAll(tempFilesDir, 0o755)

	if len(c.spec) > 0 {

		sp := new(StateFiles)
		if err := sp.specFromYAML(c.spec); err != nil {
			return fmt.Errorf("error parsing spec file: %w", err)
		}

		for _, val := range sp.StateFiles {
			fo := fileOption{}
			fo.name = val.Path
			if err := isValidFile(fo.name, validManifestFiles); err != nil {
				return fmt.Errorf("invalid -spec file: %w", err)
			}
			c.files = append(c.files, fo)
		}
		c.files.sort()
	}

	// read the TOML/YAML desired state file
	for _, f := range c.files {
		var fileState state

		if err := fileState.fromFile(f.name); err != nil {
			return err
		}

		log.Infof("Parsed [[ %s ]] successfully and found [ %d ] apps", f.name, len(fileState.Apps))
		// Merge Apps that already existed in the state
		for appName, app := range fileState.Apps {
			if _, ok := s.Apps[appName]; ok {
				if err := mergo.Merge(s.Apps[appName], app, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
					return fmt.Errorf("failed to merge %s from desired state file %s: %w", appName, f.name, err)
				}
			}
		}

		// Merge the remaining Apps
		if err := mergo.Merge(&s.Apps, &fileState.Apps); err != nil {
			return fmt.Errorf("failed to merge desired state file %s: %w", f.name, err)
		}
		// All the apps are already merged, make fileState.Apps empty to avoid conflicts in the final merge
		fileState.Apps = make(map[string]*release)

		if err := mergo.Merge(s, &fileState, mergo.WithAppendSlice, mergo.WithOverride); err != nil {
			return fmt.Errorf("failed to merge desired state file %s: %w", f.name, err)
		}
	}

	s.init() // Set defaults
	s.disableUntargetedApps(c.group, c.target)

	if !c.skipValidation {
		// validate the desired state content
		if len(c.files) > 0 {
			log.Info("Validating desired state definition")
			if err := s.validate(); err != nil { // syntax validation
				return err
			}
		}
	} else {
		log.Info("Desired state validation is skipped.")
	}

	if c.debug {
		s.print()
	}
	return nil
}

// getRunFlags returns dry-run and debug flags
func (c *cli) getRunFlags() []string {
	if c.dryRun {
		return []string{"--dry-run", "--debug"}
	}
	if c.debug {
		return []string{"--debug"}
	}
	return []string{}
}
