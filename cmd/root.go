package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/panekj/mailru-dl/pkg/types"
)

const version = "v0.2.0-dev"

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
	log   = logrus.New()
	regex = regexp.MustCompile(`cloud\.mail\.ru/public/(.+)`)
	c     = http.Client{
		Jar: &cookiejar.Jar{},
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}
	conf = &flags{}
)

type flags struct {
	wait     time.Duration
	workDir  string
	logLevel string
	prefix   bool
	retry    bool
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.Flags().StringVarP(&conf.logLevel, "log-level", "l", "info", "Log level")
	rootCmd.Flags().StringVarP(&conf.workDir, "workdir", "w", ".", "Download path")
	rootCmd.Flags().DurationVarP(&conf.wait, "wait", "d", time.Second*5, "Wait time before requests")
	rootCmd.Flags().BoolVar(&conf.prefix, "prefix", false, "Add unique prefix path to avoid file collision")
	rootCmd.Flags().BoolVar(&conf.retry, "retry", true, "Automatically retries download if failed")
}

// TODO:
//      - add downloads async
//      - add progress bar
func down(args []string) {
	log.SetFormatter(&logrus.TextFormatter{
		DisableLevelTruncation: true,
		PadLevelText:           true,
	})
	log.SetOutput(colorable.NewColorableStdout())

	if l, err := logrus.ParseLevel(conf.logLevel); err != nil {
		log.Panic(err)
	} else {
		log.Infof("Log level set to '%s' and '%s' was requested", l.String(), conf.logLevel)
		log.SetLevel(l)
	}

	for _, link := range args {
		link := strings.TrimSuffix(link, `/`)

		if m := regex.FindAllStringSubmatch(link, -1); m != nil {
			log.Debugf("%+v", m)
			link = m[0][1]
		} else {
			log.Errorf("Invalid URL: %s", link)
		}

		wd := conf.workDir
		if conf.prefix {
			l := strings.Split(link, `/`)
			wd = filepath.Join(wd, l[0], l[1])
		}

		recurseDownload(link, wd, 0)
	}
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

	if err := os.MkdirAll(path, 0777); err != nil {
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
		log.Infof("Sleeping for %d", conf.wait)
		time.Sleep(conf.wait)
	}
}

func fileDownload(name, url, weblink string, size int64) {
start:
	log.Debug(weblink)
	log.Debug(name)

	info, err := c.Head(fmt.Sprintf("%s/%s", url, weblink))
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Stat(name)
	if errors.Is(err, fs.ErrNotExist) {
		log.Debug(err)
	} else {
		log.Debugf("loc: %v | rem: %v dif: %v | con: %v dif: %v", f.Size(), size, size-f.Size(), info.ContentLength, info.ContentLength-f.Size())
		if f.Size() != info.ContentLength {
			if err = os.Remove(name); err != nil {
				log.Panic(err)
			}
		} else {
			log.Infof("File %s already downloaded. Local: %v Remote: %v | Difference: %v", name, size, f.Size(), size-f.Size())
			return
		}
	}

	file, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := c.Get(fmt.Sprintf("%s/%s", url, weblink))
	if err != nil {
		log.Error(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error(err)
		}
	}(resp.Body)

	s, err := io.Copy(file, resp.Body)
	if err != nil {
		log.Error(err)
	}

	var diff int64
	if s > size {
		diff = s - size
	} else {
		diff = size - s
	}

	if s != size {
		log.Warnf("Mismatch! Local: %v Remote: %v Difference: %v", s, size, diff)
		if conf.retry {
			log.Infof("Retrying in %v...", conf.wait)
			time.Sleep(conf.wait)
			goto start
		}
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error(err)
		}
	}(file)

	log.Infof("File %s downloaded successfully with size %d", name, s)
}
