package handler

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
)

var (
	flags = struct {
		owner       string
		repos       flagutil.StringSliceFlag
		accessToken string
		fixLabels   bool
		dryRun      bool
	}{}

	// TODO(jonboulle): really we could just use the map now since we don't
	// even really use these objects. Although the sorting is nice.
	requiredLabels = []github.Label{
		github.Label{Color: sp("bfe5bf"), Name: sp("area/developer tooling")},
		github.Label{Color: sp("bfe5bf"), Name: sp("area/distribution")},
	}
	// map of label name -> color
	requiredLabelsMap = map[string]string{}

	// check for existing labels that look like close matches of the desired ones
	// (will be checked case insensitively)
	fuzzyMatches = map[string][]string{
		//		"area/developer tooling": {"build", "scripts", "tool", "tooling", "tools"},
		"priority/P1":           {"priority1", "p1"},
		"kind/bug":              {"bug"},
		"kind/regression":       {"regression"},
		"kind/release":          {"release"},
		"kind/question":         {"question"},
		"kind/enhancement":      {"enhancement", "feature request", "request"},
		"kind/support":          {"help", "support"},
		"reviewed/duplicate":    {"duplicate", "dup"},
		"reviewed/needs rebase": {"needs rebase"},
		"reviewed/needs rework": {"waiting on customer"},
		"reviewed/needs tests":  {"needs tests"},
		"reviewed/won't fix":    {"invalid", "wontfix", "won'tfix", "won't fix"},
	}
	// reverse map of fuzzy label names -> required label names
	fuzzyMatchesMap = map[string]string{}
)

func init() {
	for _, lbl := range requiredLabels {
		if _, ok := requiredLabelsMap[*lbl.Name]; ok {
			panic(fmt.Sprintf("set %s twice", *lbl.Name))
		}
		requiredLabelsMap[*lbl.Name] = *lbl.Color
	}
	for k, v := range fuzzyMatches {
		for _, s := range v {
			if _, ok := fuzzyMatchesMap[s]; ok {
				panic(fmt.Sprintf("set %s twice", s))
			}
			fuzzyMatchesMap[s] = k
		}
	}
}

func sp(s string) *string {
	return &s
}

func init() {
	flag.StringVar(&flags.owner, "owner", "coreos", "owner of the repositories to be checked")
	flag.Var(&flags.repos, "repos", "repositories to be checked; if empty, all repositories of the given owner will be checked")
	flag.StringVar(&flags.accessToken, "access-token", "", "the GitHub API access token")
	flag.BoolVar(&flags.fixLabels, "fix", false, "fix labels (default behaviour is just to check)")
	flag.BoolVar(&flags.dryRun, "dry-run", true, "print actions that would be taken")
}

func main() {
	os.Exit(2)

	flag.Parse()

	client := github.NewClient(
		func(accessToken string) *http.Client {
			if accessToken == "" {
				return nil
			}
			return oauth2.NewClient(
				oauth2.NoContext,
				oauth2.StaticTokenSource(
					&oauth2.Token{
						AccessToken: flags.accessToken,
					},
				),
			)
		}(flags.accessToken),
	)
	repos := []string(flags.repos)
	if len(repos) < 1 {
		fmt.Println("Fetching all repositories for", flags.owner)
		rr, _, err := client.Repositories.List(
			flags.owner,
			&github.RepositoryListOptions{
				ListOptions: github.ListOptions{
					PerPage: 1000,
				},
			},
		)
		if err != nil {
			log.Fatalf("failed to fetch repositories: %v", err)
		}
		for _, r := range rr {
			repos = append(repos, *r.Name)
		}
	}
	exit := 0
	for _, repo := range repos {
		svc := client.Issues
		fmt.Printf("Analysing repo %s/%s\n", flags.owner, repo)
		lbls, _, err := svc.ListLabels(flags.owner, repo, &github.ListOptions{PerPage: 1000})
		if err != nil {
			log.Fatalf("failed to list labels: %v", err)
		}

		hasLbls := map[string]string{}
		for _, lbl := range lbls {
			hasLbls[*lbl.Name] = *lbl.Color
		}

		missingLbls := []string{}
		badLbls := []string{}
		fmt.Println("  Checking required labels")
		for _, reqLbl := range requiredLabels {
			reqLbl := reqLbl
			name := *reqLbl.Name
			color, ok := hasLbls[name]
			if !ok {
				fmt.Printf("    missing label: %s\n", name)
				missingLbls = append(missingLbls, name)
			} else if strings.ToLower(color) != strings.ToLower(*reqLbl.Color) {
				fmt.Printf("    bad coloUred label (got %v, want %v): %s\n", color, *reqLbl.Color, name)
				badLbls = append(badLbls, name)
			}
		}

		if !flags.fixLabels {
			if len(missingLbls) > 0 || len(badLbls) > 0 {
				exit = 1
			}
			continue
		}

		migrateLabels := map[string]string{}
		for lbl, _ := range hasLbls {
			if wlbl, ok := fuzzyMatchesMap[strings.ToLower(lbl)]; ok {
				migrateLabels[lbl] = wlbl
			}
		}

		if len(missingLbls) > 0 {
			fmt.Println("  Creating missing labels")
			if flags.dryRun {
				fmt.Println("    ****dry run****")
			}
			for _, lbl := range missingLbls {
				fmt.Printf("    creating label: %v\n", lbl)
				wlbl := &github.Label{Name: sp(lbl), Color: sp(requiredLabelsMap[lbl])}
				if !flags.dryRun {
					_, _, err := svc.CreateLabel(flags.owner, repo, wlbl)
					if err != nil {
						log.Fatalf("failed to create label: %v", err)
					}
				}
			}
		}

		if len(badLbls) > 0 {
			fmt.Println("  Fixing bad labels")
			if flags.dryRun {
				fmt.Println("    ****dry run****")
			}
			for _, lbl := range badLbls {
				fmt.Printf("    fixing label: %v\n", lbl)
				wlbl := &github.Label{Name: sp(lbl), Color: sp(requiredLabelsMap[lbl])}
				if !flags.dryRun {
					_, _, err := svc.EditLabel(flags.owner, repo, lbl, wlbl)
					if err != nil {
						log.Fatalf("failed to fix label: %v", err)
					}
				}
			}
		}

		if len(migrateLabels) > 0 {
			fmt.Println("  Migrating similar labels")
			if flags.dryRun {
				fmt.Println("    ****dry run****")
			}
			migrateLabel := func(from, to string) {
				issues, _, err := svc.ListByRepo(flags.owner, repo, &github.IssueListByRepoOptions{
					Labels: []string{from},
					ListOptions: github.ListOptions{
						PerPage: 100000,
					},
				})
				if err != nil {
					log.Fatalf("error listing issues: %v", err)
				}
				for _, issue := range issues {
					n := *issue.Number
					_, _, err = svc.AddLabelsToIssue(flags.owner, repo, n, []string{to})
					if err != nil {
						log.Fatalf("error adding label: %v", err)
					}
					_, err = svc.RemoveLabelForIssue(flags.owner, repo, n, from)
					if err != nil {
						log.Fatalf("error removing label from issue: %v", err)
					}
				}
				if _, err = svc.DeleteLabel(flags.owner, repo, from); err != nil {
					log.Fatalf("error deleting label: %v", err)
				}
			}
			for from, to := range migrateLabels {
				fmt.Printf("    migrating label: %v to %v\n", from, to)
				if !flags.dryRun {
					migrateLabel(from, to)
				}
			}
		}
	}
	fmt.Println("Finished.")
	os.Exit(exit)
}
