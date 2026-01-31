package models

import "time"

// State represents the state.yaml file structure for session state
type State struct {
	LastUpdated   time.Time `yaml:"last_updated"`
	Goal          string    `yaml:"goal,omitempty"`
	Progress      string    `yaml:"progress,omitempty"`
	Blocker       string    `yaml:"blocker,omitempty"`
	Next          []string  `yaml:"next,omitempty"`
	WorkingFiles  []string  `yaml:"working_files,omitempty"`
	OpenQuestions []string  `yaml:"open_questions,omitempty"`
}

// NewState creates a new empty state
func NewState() *State {
	return &State{
		LastUpdated: time.Now(),
		Next:        []string{},
	}
}

// Update updates the state with new values
func (s *State) Update(goal, progress, blocker string, next []string) {
	s.LastUpdated = time.Now()
	if goal != "" {
		s.Goal = goal
	}
	if progress != "" {
		s.Progress = progress
	}
	s.Blocker = blocker // Allow clearing blocker
	if len(next) > 0 {
		s.Next = next
	}
}

// AddWorkingFile adds a file to the working files list
func (s *State) AddWorkingFile(file string) {
	for _, f := range s.WorkingFiles {
		if f == file {
			return
		}
	}
	s.WorkingFiles = append(s.WorkingFiles, file)
}

// AddOpenQuestion adds a question to the open questions list
func (s *State) AddOpenQuestion(question string) {
	for _, q := range s.OpenQuestions {
		if q == question {
			return
		}
	}
	s.OpenQuestions = append(s.OpenQuestions, question)
}
