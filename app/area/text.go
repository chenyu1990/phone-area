package area

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"phone-area/schema"
	"regexp"
	"sync"
)

type Text struct {
	File        string
	ProjectFile string
	OutputFile  string
}

func NewText(file, outputFile, projectFile string) *Text {
	return &Text{
		File:        file,
		ProjectFile: projectFile,
		OutputFile:  outputFile,
	}
}

func (a *Text) Run() error {
	phonesInfo, err := a.getPhones()
	if err != nil {
		return err
	}

	projects, err := a.getProjects()
	if err != nil {
		return err
	}

	group := sync.WaitGroup{}
	for _, info := range phonesInfo {
		group.Add(1)
		go func(pi *schema.PhoneInfo) {
			res, err := getInfo(pi)
			if err != nil {
				return
			}

			pi.Province = res.Data.Province
			pi.City = res.Data.City
			pi.Area = pi.Province + pi.City
			pi.ServiceProvider = res.Data.ServiceProvider
			group.Done()
		}(info)
	}
	group.Wait()

	of, err := os.Create(a.OutputFile)
	if err != nil {
		return err
	}
	defer of.Close()

	for _, info := range phonesInfo {
		if o, ok := projects[info.WebSite]; ok {
			info.Project = o
		}
		_, err := of.WriteString(fmt.Sprintf(
			"%s\t%s\t%s\t%s\t%s\t%s\t\n",
			info.Number,
			info.Project,
			info.Province,
			info.City,
			info.ServiceProvider,
			info.WebSite,
		))
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *Text) getProjects() (map[string]string, error) {
	projects := make(map[string]string)
	if a.ProjectFile == "" {
		return projects, nil
	}

	fi, err := os.Open(a.ProjectFile)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		str := string(a)
		if str == "" {
			continue
		}

		compile := regexp.MustCompile("(\\S+)\\s+(\\S+)")
		subMatch := compile.FindStringSubmatch(str)
		if len(subMatch) > 0 {
			webSite := subMatch[1]
			project := subMatch[2]
			projects[webSite] = project
		}

	}

	return projects, nil
}

func (a *Text) getPhones() (schema.PhoneInfos, error) {
	fi, err := os.Open(a.File)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	phoneInfos := schema.PhoneInfos{}

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		number := string(a)
		if number == "" {
			continue
		}

		phoneInfo := &schema.PhoneInfo{
		}
		compile := regexp.MustCompile("[0-9]{11}")
		subMatch := compile.FindStringSubmatch(number)
		if len(subMatch) > 0 {
			phoneInfo.Number = subMatch[0]
		}

		compile = regexp.MustCompile("https?://\\S*")
		subMatch = compile.FindStringSubmatch(number)
		if len(subMatch) > 0 {
			phoneInfo.WebSite = subMatch[0]
		}
		phoneInfos = append(phoneInfos, phoneInfo)
	}

	return phoneInfos, nil
}
