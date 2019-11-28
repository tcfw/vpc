package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/tcfw/vpc/pkg/vpc"
)

//NewMigrateCmd provides a command to update schema
func NewMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Runs all schema migrations",
		Run: func(cmd *cobra.Command, args []string) {
			db, err := vpc.DBConn()
			if err != nil {
				log.Fatalf("Failed to open DB: %s", err)
			}

			path, _ := cmd.Flags().GetString("dir")

			if err := vpc.Migrate(db, path); err != nil {
				log.Fatalln(err)
			}
		},
	}

	cmd.Flags().StringP("dir", "d", "./migrations", "Directory of migrations to run")

	return cmd
}
