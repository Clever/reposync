// reposync syncs repos for a GitHub user into a folder on your computer.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/rgarcia/reposync/sticky"
	"golang.org/x/oauth2"
)

func main() {
	user := flag.String("user", "", "Git Hub user or organization you'd like to sync a folder with")
	dir := flag.String("dir", "", "Directory to put folders for each repo")
	archivedir := flag.String("archivedir", "", "Directory to move folders in dir that are not associated with a repo")
	token := flag.String("token", "", "GitHub token to use for auth")
	dryrun := flag.Bool("dryrun", false, "Set to true to print actions instead of performing them")
	flag.Parse()
	if *user == "" {
		log.Fatal("must provide user")
	}
	if *dir == "" {
		log.Fatal("must provide dir")
	}
	if *archivedir == "" {
		log.Fatal("must provide archivedir")
	}
	if *token == "" {
		log.Fatal("must provide token")
	}
	rs := RepoSync{
		org:        *user,
		workdir:    *dir,
		archivedir: *archivedir,
		token:      *token,
		dryrun:     *dryrun,
	}
	if err := rs.Sync(); err != nil {
		log.Fatal(err)
	}
}

// Contains computes whether an element is a member of a set of strings.
func Contains(list []string, el string) bool {
	for _, str := range list {
		if str == el {
			return true
		}
	}
	return false
}

// Difference computes the set difference A - B for string sets.
func Difference(a, b []string) []string {
	diff := []string{}
	for _, str := range a {
		if Contains(b, str) {
			continue
		}
		diff = append(diff, str)
	}
	return diff
}

// TaskWithReporter runs a function and reports on its status via sticky.Line.
type TaskWithReporter struct {
	task        func() error
	reporter    sticky.Line
	description string
}

func NewTaskWithReporter(task func() error, reporter sticky.Line, description string) *TaskWithReporter {
	return &TaskWithReporter{task: task, reporter: reporter, description: description}
}

func (tws *TaskWithReporter) Run() {
	result := make(chan error)
	go func() {
		result <- tws.task()
	}()
	status := sticky.NewStatusPart()
	descriptionWithSpace := " " + tws.description
	for {
		select {
		case <-time.Tick(1 * time.Second):
			status.Active()
			tws.reporter.DisplayP(status, sticky.NewTextPart(descriptionWithSpace).Color(color.FgYellow, color.Bold))
		case err := <-result:
			if err != nil {
				status.Fail()
				tws.reporter.DisplayP(status, sticky.NewTextPart(fmt.Sprintf("%s: %s", descriptionWithSpace, err)).Color(color.FgRed, color.Bold))
			} else {
				tws.reporter.DisplayP(status, sticky.NewTextPart(descriptionWithSpace).Color(color.FgGreen, color.Bold))
				status.Success()
			}
			return
		}
	}
}

type RepoSync struct {
	org         string
	workdir     string
	archivedir  string
	token       string
	language    string
	languageNot string
	dryrun      bool
}

func (rs RepoSync) Sync() error {

	// get list of repos for org
	var allRepos []string
	NewTaskWithReporter(func() error {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: rs.token},
		)
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		client := github.NewClient(tc)
		opt := &github.RepositoryListByOrgOptions{
			Type:        "all",
			ListOptions: github.ListOptions{PerPage: 100},
		}
		for {
			repos, resp, err := client.Repositories.ListByOrg(rs.org, opt)
			if err != nil {
				return err
			}
			// add repos that filters (if any)
			for _, repo := range repos {
				if rs.language != "" && (repo.Language == nil || *repo.Language != rs.language) {
					continue
				}
				if rs.languageNot != "" && !(repo.Language == nil || *repo.Language != rs.languageNot) {
					continue
				}
				if repo.Name == nil {
					continue
				}
				allRepos = append(allRepos, *repo.Name)
			}
			if resp.NextPage == 0 {
				break
			}
			opt.ListOptions.Page = resp.NextPage
		}
		return nil
	}, sticky.NewBlock(1).Line(0), fmt.Sprintf("loading repos for %s", rs.org)).Run()

	// get list of current repositories checked out, ignoring non-directories and hidden directories
	var currentRepos []string
	NewTaskWithReporter(func() error {
		files, _ := ioutil.ReadDir(rs.workdir)
		for _, f := range files {
			if !f.IsDir() || strings.Index(f.Name(), ".") == 0 {
				continue
			}
			currentRepos = append(currentRepos, f.Name())
		}
		return nil
	}, sticky.NewBlock(1).Line(0), fmt.Sprintf("loading repos already cloned in %s", rs.workdir)).Run()

	reposToArchive := Difference(currentRepos, allRepos)
	reposToClone := Difference(allRepos, currentRepos)

	if len(reposToArchive)+len(reposToClone) == 0 {
		sticky.NewBlock(1).Line(0).DisplayP(sticky.NewStatusPart().Success(), sticky.NewTextPart(" nothing to do!").Color(color.FgGreen, color.Bold))
		return nil
	}

	block := sticky.NewBlock(len(reposToArchive) + len(reposToClone))

	var archivers sync.WaitGroup
	if err := os.MkdirAll(rs.archivedir, 0755); err != nil {
		return err
	}
	for idx, repo := range reposToArchive {
		archivers.Add(1)
		i := idx
		r := repo
		go func() {
			defer archivers.Done()
			NewTaskWithReporter(func() error {
				if rs.dryrun {
					return nil
				}
				return os.Rename(path.Join(rs.workdir, r), path.Join(rs.archivedir, r))
			}, block.Line(i), fmt.Sprintf("archiving %s", r)).Run()
		}()
	}

	var cloners sync.WaitGroup
	for idx, repo := range reposToClone {
		cloners.Add(1)
		i := idx
		r := repo
		go func() {
			defer cloners.Done()
			NewTaskWithReporter(func() error {
				if rs.dryrun {
					return nil
				}
				return exec.Command("git", "clone", fmt.Sprintf("git@github.com:%s/%s", rs.org, r), path.Join(rs.workdir, r)).Run()
			}, block.Line(i+len(reposToArchive)), fmt.Sprintf("cloning %s", r)).Run()
		}()
	}

	archivers.Wait()
	cloners.Wait()
	return nil
}
