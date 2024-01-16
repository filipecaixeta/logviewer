package fake

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/filipecaixeta/logviewer/internal/config"
	"github.com/filipecaixeta/logviewer/internal/source"
	"github.com/filipecaixeta/logviewer/internal/source/k8s"
	"github.com/filipecaixeta/logviewer/internal/state"

	tea "github.com/charmbracelet/bubbletea"
)

type Fake struct {
	columns []*source.List
}

func New(config *config.Config) source.Source {
	f := &Fake{}
	r := rand.New(rand.NewSource(1))
	namespaces := []*k8s.Namespace{}
	for _, namespace := range config.Namespaces {
		n := &k8s.Namespace{
			Name: namespace,
		}
		n.Workloads = []*k8s.Workload{}
		// generate random numer of workloads
		for i := 0; i < r.Int()%4+1; i++ {
			w := &k8s.Workload{
				Name: fmt.Sprintf("workload-%d", i),
			}
			for j := 0; j < r.Int()%5+2; j++ {
				p := &k8s.Pod{
					Name: fmt.Sprintf("pod-%d", r.Int()),
				}
				for k := 0; k < r.Int()%2+1; k++ {
					c := &k8s.Container{
						Name: fmt.Sprintf("container-%d", k),
					}
					p.Containers = append(p.Containers, c)
				}
				w.Pods = append(w.Pods, p)
			}
			n.Workloads = append(n.Workloads, w)
		}
		namespaces = append(namespaces, n)
	}

	f.columns = []*source.List{
		source.NewList("Namespaces", source.ConvertInterface2ListItems(namespaces)),
		source.NewList("Workloads", []source.ListItem{}),
		source.NewList("Pods", []source.ListItem{}),
		source.NewList("Containers", []source.ListItem{}),
	}

	return f
}

func (f *Fake) Init(stateChan chan state.State) tea.Cmd {
	return func() tea.Msg {
		stateChan <- state.StateLoading
		time.Sleep(100 * time.Millisecond)
		stateChan <- state.StateBrose
		return nil
	}
}

func (f *Fake) Columns() []*source.List {
	return f.columns
}

func (f *Fake) getLogCfg() map[string]string {
	cfg := map[string]string{}
	for _, c := range f.columns {
		s := c.SelectedItem()
		cfg[strings.ToLower(c.Title)] = s.String()
		if c.IsActive() {
			break
		}
	}
	return cfg
}

func (f *Fake) Logs(ctx context.Context, stateChan chan state.State, logChan chan string) tea.Cmd {
	cfg := f.getLogCfg()
	return func() tea.Msg {
		stateChan <- state.StateLogs

		v := map[string]interface{}{
			"level":   "info",
			"message": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Curabitur eget rutrum libero. Nunc malesuada tincidunt urna ac auctor. Sed lacinia orci vitae neque vulputate, tempus ullamcorper urna mollis. Phasellus consectetur lorem nisi, iaculis vestibulum libero luctus vel. Etiam sit amet gravida nulla, ac gravida lacus. Nunc dolor nunc, bibendum sed turpis eget, condimentum euismod libero. Duis eu velit sit amet enim vestibulum dictum ut ut metus. Integer interdum vitae lorem at finibus. Nullam condimentum in dolor suscipit pulvinar. Phasellus nunc felis, sagittis eu tortor et, elementum varius quam. Quisque auctor sapien dolor, et molestie magna porttitor a. Mauris rutrum, nisi et condimentum pulvinar, metus metus faucibus purus, eget tincidunt lectus quam sit amet arcu. Vivamus et erat lobortis, luctus lorem sit amet, elementum ligula. Nam ut diam viverra, venenatis sapien vitae, luctus est. Phasellus ullamcorper facilisis consectetur. Duis porta diam in mi consequat porttitor.",
		}
		jsonLog, _ := json.Marshal(v)

		for _, l := range []string{
			string(jsonLog),
			"192.168.1.101 - - [22/Oct/2021:13:14:15 +0000] \"GET /index.html HTTP/1.1\" 200 2326 \"http://www.google.com\" \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537",
			"192.168.1.102 - - [22/Oct/2021:13:15:16 +0000] \"POST /contact.html HTTP/1.1\" 404 154 \"http://www.bing.com\" \"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:54.0) Gecko/20100101 Firefox/54.0",
			"192.168.1.103 - - [22/Oct/2021:13:16:17 +0000] \"PUT /products.html HTTP/1.1\" 500 1234 \"-\" \"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; AS; rv:11.0) like Gecko\"",
			"192.168.1.104 - - [22/Oct/2021:13:17:18 +0000] \"DELETE /services.html HTTP/1.1\" 200 5678 \"http://www.yahoo.com\" \"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537",
			"192.168.1.105 - - [22/Oct/2021:13:18:19 +0000] \"GET /about.html HTTP/1.1\" 404 9012 \"-\" \"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:54.0) Gecko/20100101 Firefox/54.0",
		} {
			select {
			case <-ctx.Done():
				return nil
			case logChan <- l:
			}
		}

		for i := 0; i < 500; i++ {
			item := map[string]interface{}{
				"level":       "info",
				"msg":         fmt.Sprintf("this is a test log %d", i),
				"ts":          float64(time.Now().UnixNano()) / 1e9,
				"int_field":   int64(i),
				"bool_field":  i%2 == 0,
				"float_field": float64(i) + 0.5,
			}
			for k, v := range cfg {
				item[k] = v
			}

			jsonLog, _ := json.Marshal(item)
			select {
			case <-ctx.Done():
				return nil
			case logChan <- string(jsonLog):
				time.Sleep(100 * time.Millisecond)
			}
		}
		return nil
	}
}

func (f *Fake) Close() {
}
