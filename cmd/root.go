package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/panekj/mailru-dl/pkg/types"
)

const version = "v0.0.0-dev"

var rootCmd = &cobra.Command{
	Use:     "mailru-dl",
	Short:   "dl dir from cloud.mail.ru",
	Long:    "",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		down(args)
	},
}

var (
	log    = logrus.New()
	cwd, _ = os.Getwd()
	regex  = regexp.MustCompile(`cloud\.mail\.ru/public/(.+)`)
	c      = http.Client{
		Jar: &cookiejar.Jar{},
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
)

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize()
	logLevel := rootCmd.Flags().String("l", "info", "log level")

	log.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: false,
	})

	if l, err := logrus.ParseLevel(*logLevel); err == nil {
		log.SetLevel(l)
	} else {
		logrus.Warn(err)
	}
}

// TODO:
//      - add downloads async
//      - add progress bar
//      - add download path option
func down(args []string) {
	link := strings.TrimSuffix(args[0], `/`)

	if m := regex.FindAllStringSubmatch(link, -1); m != nil {
		link = m[0][1]
	} else {
		log.Panic("no matches")
	}

	recurseDownload(link, "", 0)
}

func get(q string, values url.Values) []byte {
	requestURL := types.EndpointURL + "/" + q + "?" + values.Encode()

	log.Debug(requestURL)

	r, err := c.Get(requestURL)
	if err != nil {
		log.Error(r)
		log.Panic(err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error(err)
		}
	}(r.Body)

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Panic(err)
	}

	if r.StatusCode != http.StatusOK {
		log.Errorf("code: %d, resp: %s", r.StatusCode, string(b))
		return nil
	}

	return b
}

func recurseDownload(weblink, path string, limit int) {
	log.Debug(weblink)
	values := url.Values{
		"weblink": {weblink},
		"sort":    {`{"type":"name","order":"asc"}`},
		"offset":  {"0"},
		"limit":   {fmt.Sprint(limit)},
		"api":     {fmt.Sprint(types.APIVersion)},
		"build":   {types.Build},
	}

	log.Debugf("%v", values)

	var b []byte
	var r types.Response

	if limit == 0 {
		b := get("folder", values)

		if err := json.Unmarshal(b, &r); err != nil {
			log.Panic(err)
		}

		values["limit"] = []string{fmt.Sprint(r.Body.Count.Files + r.Body.Count.Folders)}
	}

	b = get("folder", values)
	r = types.Response{}
	if err := json.Unmarshal(b, &r); err != nil {
		log.Panic(err)
	}

	path = filepath.Join(path, r.Body.Name)

	if err := os.MkdirAll(filepath.Join(cwd, path), 0777); err != nil {
		log.Panic(err)
	}

	for _, v := range r.Body.List {
		if v.Type == "file" {
			b := get("dispatcher", url.Values{
				"api":   {fmt.Sprint(types.APIVersion)},
				"build": {types.Build},
				"_":     {fmt.Sprint(time.Since(time.Unix(0, 0)).Milliseconds())},
			})

			var r types.Response
			if err := json.Unmarshal(b, &r); err != nil {
				log.Panic(err)
			}

			fileDownload(filepath.Join(path, v.Name), r.Body.WeblinkGet[0].URL, v.Weblink, v.Size)
		}
		if v.Type == "folder" {
			recurseDownload(v.Weblink, path, v.Count.Folders+v.Count.Files)
		}
	}
}

func fileDownload(name, url, weblink string, size int64) {
	log.Debug(weblink)
	log.Debug(name)

	f, err := os.Stat(name)
	if os.IsExist(err) {
		if f.Size() < size {
			log.Warnf("loc: %v rem: %v", f.Size(), size)
			if err = os.Remove(name); err != nil {
				log.Panic(err)
			}
		} else {
			log.Infof("File %s already downloaded. lsize: %v rsize: %v", name, size, f.Size())
			return
		}
	}

	file, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.Get(fmt.Sprintf("%s/%s", url, weblink))
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error(err)
		}
	}(resp.Body)

	s, err := io.Copy(file, resp.Body)
	if err != nil {
		log.Panic(err)
	}

	if s != size {
		log.Warnf("Mismatch! Local: %v Remote: %v", s, size)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error(err)
		}
	}(file)

	log.Infof("file: %s size: %d", name, s)
}
