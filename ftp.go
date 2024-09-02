package main

import (
	"fmt"
	filedriver "github.com/goftp/file-driver"
	"github.com/goftp/server"
	"github.com/urfave/cli/v2"
	"strconv"
)

func FTPServerCommand(app *cli.App) {
	var (
		hostName   string
		listenPort int
		err        error
		userName   string
		password   string
		rootPath   string
	)

	command := cli.Command{
		Name:    "ftp",
		Aliases: nil,
		Usage: `simulates an ftp server
		cli args -> ftp <hostname> <listenport>`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "username",
				Usage:       "Username for the FTP server auth.",
				Required:    false,
				Value:       "",
				Destination: &userName,
			},
			&cli.StringFlag{
				Name:        "password",
				Usage:       "Password for the FTP server auth.",
				Required:    false,
				Value:       "",
				Destination: &password,
			},
			&cli.StringFlag{
				Name:        "rootpath",
				Usage:       "The root path of the FTP server. For the current directory use it with ./",
				Required:    false,
				Value:       "",
				Destination: &rootPath,
			},
		},
		Action: func(c *cli.Context) error {
			args := c.Args()

			hostName = args.Get(0)

			listenPort, err = strconv.Atoi(args.Get(1))
			if err != nil {
				return fmt.Errorf("invalid port number. Can not parse string to int: %w", err)
			}

			ftpServer := server.NewServer(&server.ServerOpts{
				Factory: &filedriver.FileDriverFactory{
					RootPath: rootPath,
					Perm:     server.NewSimplePerm("user", "group"),
				},
				Port:     listenPort,
				Hostname: hostName,
				Auth:     &server.SimpleAuth{Name: userName, Password: password},
			})
			err := ftpServer.ListenAndServe()
			if err != nil {
				return fmt.Errorf("starting FTP server failed: %w", err)
			}

			return nil
		},
		Subcommands: nil,
	}

	app.Commands = append(app.Commands, &command)

}
