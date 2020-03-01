package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/logrusorgru/aurora"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	api "github.com/textileio/textile/api/client"
	"github.com/textileio/textile/api/pb"
	"github.com/textileio/textile/cmd"
)

func init() {
	rootCmd.AddCommand(projectsCmd)
	rootCmd.AddCommand(initProjectCmd) // alias to root
	projectsCmd.AddCommand(initProjectCmd, lsProjectsCmd, inspectProjectCmd, rmProjectCmd, shellProjectCmd)

	initProjectCmd.Flags().String(
		"path",
		".",
		"Project path")
}

var projectsCmd = &cobra.Command{
	Use: "projects",
	Aliases: []string{
		"project",
	},
	Short: "Manage projects",
	Long:  `Manage your projects.`,
	Run: func(c *cobra.Command, args []string) {
		lsProjects()
	},
}

var initProjectCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "Init a new project",
	Long:  `Initialize a new project.`,
	Args:  cobra.ExactArgs(1),
	Run: func(c *cobra.Command, args []string) {
		var pth string
		if !c.Flag("path").Changed {
			var err error
			pth, err = os.Getwd()
			if err != nil {
				cmd.Fatal(err)
			}
		} else {
			pth = c.Flag("path").Value.String()
		}

		pth = filepath.Join(pth, ".textile")
		if err := os.MkdirAll(pth, os.ModePerm); err != nil {
			cmd.Fatal(err)
		}
		filename := filepath.Join(pth, "config.yml")

		if _, err := os.Stat(filename); err == nil {
			cmd.Fatal(fmt.Errorf("project already exists in %s", pth))
		}

		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()
		proj, err := client.AddProject(
			ctx,
			args[0],
			api.Auth{
				Token: authViper.GetString("token"),
			})
		if err != nil {
			cmd.Fatal(err)
		}
		configViper.Set("project", proj.Name)

		if err := configViper.WriteConfigAs(filename); err != nil {
			cmd.Fatal(err)
		}

		cmd.Success("Initialized empty project in %s", aurora.White(pth).Bold())
	},
}

var lsProjectsCmd = &cobra.Command{
	Use: "ls",
	Aliases: []string{
		"list",
	},
	Short: "List projects",
	Long:  `List existing projects under the current scope.`,
	Run: func(c *cobra.Command, args []string) {
		lsProjects()
	},
}

func lsProjects() {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	projects, err := client.ListProjects(
		ctx,
		api.Auth{
			Token: authViper.GetString("token"),
		})
	if err != nil {
		cmd.Fatal(err)
	}

	if len(projects.List) > 0 {
		data := make([][]string, len(projects.List))
		for i, p := range projects.List {
			data[i] = []string{p.Name, p.ID, p.StoreID}
		}
		cmd.RenderTable([]string{"name", "id", "store"}, data)
	}

	cmd.Message("Found %d projects for current scope", aurora.White(len(projects.List)).Bold())
}

var inspectProjectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Display project information",
	Long:  `Display detailed information about a project (interactive).`,
	Run: func(c *cobra.Command, args []string) {
		selected := selectProject("Select project", aurora.Sprintf(
			aurora.BrightBlack("> Selected {{ .Name | white | bold }}")))

		cmd.RenderTable(
			[]string{"name", "id", "store", "address", "balance"},
			[][]string{{selected.Name, selected.ID, selected.StoreID, selected.WalletAddress,
				strconv.FormatInt(selected.WalletBalance, 10)}})
	},
}

var shellProjectCmd = &cobra.Command{
	Use:   "shell",
	Short: "Shell into a project",
	Long:  `Shel into the ThreadDB of a project (interactive).`,
	Run: func(c *cobra.Command, args []string) {
		selected := selectProject("Select project", aurora.Sprintf(
			aurora.BrightBlack("> Selected {{ .Name | white | bold }}")))
		cmd.Message(authViper.GetString("token"))
		cmd.Message(selected.StoreID)

		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		os.Chdir("/Users/eighty/github/eightysteele/thread-shell")
		mydir, err := os.Getwd()
		if err == nil {
			fmt.Println(mydir)
		}
		// Set an environment variable.
		os.Setenv("SOME_VAR", "1")

		// get project token
		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()
		tokens, err := client.ListTokens(
			ctx,
			selected.Name,
			api.Auth{
				Token: authViper.GetString("token"),
			})
		if err != nil {
			cmd.Fatal(err)
		}
		// if len(tokens.List) > 0 {
		// data := make([][]string, len(tokens.List))
		// for i, t := range tokens.List {
		// 	data[i] = []string{t}
		// }
		// }

		// Transfer stdin, stdout, and stderr to the new process
		// and also set target directory for the shell to start in.
		pa := os.ProcAttr{
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
			Dir:   mydir,
		}

		// Start up a new shell.
		// Note that we supply "login" twice.
		// -fpl means "don't prompt for PW and pass through environment."
		fmt.Print(">> Starting a new shell")
		proc, err := os.StartProcess(
			"shell",
			[]string{
				"shell",
				"--project_token", tokens.GetList()[0], // @todo: hack, how handle multiple tokens?
				"--project_db_id", selected.StoreID,
				"--project_db_name", selected.Name},
			&pa)
		if err != nil {
			panic(err)
		}

		// Wait until user exits the shell
		state, err := proc.Wait()
		if err != nil {
			panic(err)
		}

		// Chang
		os.Chdir(cwd)
		fmt.Printf("<< Exited shell: %s\n", state.String())
	},
}

var rmProjectCmd = &cobra.Command{
	Use: "rm",
	Aliases: []string{
		"remove",
	},
	Short: "Remove a project",
	Long:  `Remove a project (interactive).`,
	Run: func(c *cobra.Command, args []string) {
		selected := selectProject("Remove project", aurora.Sprintf(
			aurora.BrightBlack("> Removing project {{ .Name | white | bold }}")))

		ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
		defer cancel()
		if err := client.RemoveProject(
			ctx,
			selected.Name,
			api.Auth{
				Token: authViper.GetString("token"),
			}); err != nil {
			cmd.Fatal(err)
		}

		_ = os.RemoveAll(configViper.ConfigFileUsed())

		cmd.Success("Removed project %s", aurora.White(selected.Name).Bold())
	},
}

func selectProject(label, successMsg string) *pb.GetProjectReply {
	ctx, cancel := context.WithTimeout(context.Background(), cmdTimeout)
	defer cancel()
	projects, err := client.ListProjects(
		ctx,
		api.Auth{
			Token: authViper.GetString("token"),
		})
	if err != nil {
		cmd.Fatal(err)
	}
	if len(projects.List) == 0 {
		cmd.End("You don't have any projects!")
	}

	prompt := promptui.Select{
		Label: label,
		Items: projects.List,
		Templates: &promptui.SelectTemplates{
			Active:   fmt.Sprintf(`{{ "%s" | cyan }} {{ .Name | bold }}`, promptui.IconSelect),
			Inactive: `{{ .Name | faint }}`,
			Details:  `{{ "(ID:" | faint }} {{ .ID | faint }}{{ ")" | faint }}`,
			Selected: successMsg,
		},
	}
	index, _, err := prompt.Run()
	if err != nil {
		cmd.End("")
	}

	return projects.List[index]
}
