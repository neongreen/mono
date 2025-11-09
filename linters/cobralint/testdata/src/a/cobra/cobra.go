// Package cobra is a fake cobra package for testing
package cobra

type Command struct {
	Use   string
	Short string
	RunE  func(cmd *Command, args []string) error
}

type FlagSet struct{}

func (c *Command) Flags() *FlagSet {
	return &FlagSet{}
}

func (f *FlagSet) Bool(name string, value bool, usage string) *bool {
	return nil
}

func (f *FlagSet) BoolVar(p *bool, name string, value bool, usage string) {}

func (f *FlagSet) BoolP(name string, shorthand string, value bool, usage string) *bool {
	return nil
}

func (f *FlagSet) BoolVarP(p *bool, name string, shorthand string, value bool, usage string) {}

func (f *FlagSet) String(name string, value string, usage string) *string {
	return nil
}

func (f *FlagSet) StringVar(p *string, name string, value string, usage string) {}

func (f *FlagSet) StringP(name string, shorthand string, value string, usage string) *string {
	return nil
}

func (f *FlagSet) StringVarP(p *string, name string, shorthand string, value string, usage string) {}

func (c *Command) AddCommand(cmds ...*Command) {}
