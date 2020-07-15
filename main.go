package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"os"
	"phone-area/app/area"
)

func main() {
	app := cli.NewApp()
	app.Name = "PhoneArea"
	app.Version = "1.1.0"
	app.Usage = "手机号码归属地查询"
	app.Commands = []*cli.Command{
		TextFileCommand(),
		ExcelFileCommand(),
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TextFileCommand() *cli.Command {
	return &cli.Command{
		Name:  "txt",
		Usage: "文本文件模式",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "手机号码文件",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "输出路径",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "project",
				Aliases: []string{"p"},
				Usage:   "项目映射文件  格式：网址 项目",
			},
		},
		Action: func(c *cli.Context) error {
			return area.NewText(
				c.String("file"),
				c.String("output"),
				c.String("project"),
			).Run()
		},
	}
}

func ExcelFileCommand() *cli.Command {
	return &cli.Command{
		Name:  "excel",
		Usage: "Excel模式",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "file",
				Aliases:  []string{"f"},
				Usage:    "手机号码文件",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "输出路径",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "project",
				Aliases: []string{"p"},
				Usage:   "项目映射文件  格式：网址 项目",
			},
		},
		Action: func(c *cli.Context) error {
			return area.NewExcel(
				c.String("file"),
				c.String("output"),
				c.String("project"),
			).Run()
		},
	}
}
