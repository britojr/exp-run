package cmd

import "flag"

type Command struct {
	// Run runs the command. The args are the arguments after the command name.
	Run func(cm *Command, args []string)
	// Short is a short description
	Short string
	// Long is a long message
	Long string
	// Flag is a set of flags specific to this command
	Flag *flag.FlagSet
}
