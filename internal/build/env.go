// Copyright 2016 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package build

import (
	"flag"
	"fmt"
	"os"
)

var (
	// These flags override values in build env.
	GitCommitFlag   = flag.String("git-commit", "", `Overrides git commit hash embedded into executables`)
	GitBranchFlag   = flag.String("git-branch", "", `Overrides git branch being built`)
	GitTagFlag      = flag.String("git-tag", "", `Overrides git tag being built`)
	BuildnumFlag    = flag.String("buildnum", "", `Overrides CI build number`)
	PullRequestFlag = flag.Bool("pull-request", false, `Overrides pull request status of the build`)
)

// Environment contains metadata provided by the build environment.
type Environment struct {
	Name                string // name of the environment
	Repo                string // name of GitHub repo
	Commit, Branch, Tag string // Git info
	Buildnum            string
	IsPullRequest       bool
}

func (env Environment) String() string {
	return fmt.Sprintf("%s env (commit:%s branch:%s tag:%s buildnum:%s pr:%t)",
		env.Name, env.Commit, env.Branch, env.Tag, env.Buildnum, env.IsPullRequest)
}

// Env returns metadata about the current CI environment, falling back to LocalEnv
// if not running on CI.
func Env() Environment {
	switch {
	case os.Getenv("CI") == "true" && os.Getenv("TEAMCITY") == "true":
		return Environment{
			Name:          "teamcity",
			Repo:          os.Getenv("VCS_URL"),
			Commit:        os.Getenv("TC_COMMIT"),
			Branch:        os.Getenv("TC_BRANCH"),
			Tag:           os.Getenv("TC_TAG"),
			Buildnum:      os.Getenv("TC_BUILD_NUMBER"),
			IsPullRequest: false, //we will not build PRs.  Later
		}
	case os.Getenv("CI") == "True" && os.Getenv("APPVEYOR") == "True":
		return Environment{
			Name:          "appveyor",
			Repo:          os.Getenv("APPVEYOR_REPO_NAME"),
			Commit:        os.Getenv("APPVEYOR_REPO_COMMIT"),
			Branch:        os.Getenv("APPVEYOR_REPO_BRANCH"),
			Tag:           os.Getenv("APPVEYOR_REPO_TAG"),
			Buildnum:      os.Getenv("APPVEYOR_BUILD_NUMBER"),
			IsPullRequest: os.Getenv("APPVEYOR_PULL_REQUEST_NUMBER") != "",
		}
	default:
		return LocalEnv()
	}
}

// LocalEnv returns build environment metadata gathered from git.
func LocalEnv() Environment {
	env := applyEnvFlags(Environment{Name: "local", Repo: "ethereumproject/go-ethereum"})
	if _, err := os.Stat(".git"); err != nil {
		return env
	}
	if env.Commit == "" {
		env.Commit = RunGit("rev-parse", "HEAD")
	}
	if env.Branch == "" {
		if b := RunGit("rev-parse", "--abbrev-ref", "HEAD"); b != "HEAD" {
			env.Branch = b
		}
	}
	// Note that we don't get the current git tag. It would slow down
	// builds and isn't used by anything.
	return env
}

func applyEnvFlags(env Environment) Environment {
	if !flag.Parsed() {
		panic("you need to call flag.Parse before Env or LocalEnv")
	}
	if *GitCommitFlag != "" {
		env.Commit = *GitCommitFlag
	}
	if *GitBranchFlag != "" {
		env.Branch = *GitBranchFlag
	}
	if *GitTagFlag != "" {
		env.Tag = *GitTagFlag
	}
	if *BuildnumFlag != "" {
		env.Buildnum = *BuildnumFlag
	}
	if *PullRequestFlag {
		env.IsPullRequest = true
	}
	return env
}