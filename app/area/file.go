package area

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"phone-area/schema"
	"phone-area/util/http"
	"regexp"
	"sync"
)

type Area struct {
	File        string
	ProjectFile string
	OutputFile  string
}

type Data struct {
	Province        string `json:"province"`
	City            string `json:"city"`
	ServiceProvider string `json:"sp"`
}

type Response struct {
	Code int
	Data Data
}

func NewArea(file, outputFile, projectFile string) *Area {
	return &Area{
		File:        file,
		ProjectFile: projectFile,
		OutputFile:  outputFile,
	}
}

func (a *Area) Run() error {
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
			_ = a.getInfo(pi)
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

func (a *Area) getInfo(info *schema.PhoneInfo) error {

	body, err := http.Get(API + info.Number)
	if err != nil {
		return err
	}
	var res Response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	info.Province = res.Data.Province
	info.City = res.Data.City
	info.Area = info.Province + info.City
	info.ServiceProvider = res.Data.ServiceProvider

	return nil
}

func (a *Area) getProjects() (map[string]string, error) {
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

func (a *Area) getPhones() (schema.PhoneInfos, error) {
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
