package k8s

import (
	"bufio"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"

	"github.com/filipecaixeta/logviewer/internal/config"
	"github.com/filipecaixeta/logviewer/internal/source"
	"github.com/filipecaixeta/logviewer/internal/state"

	tea "github.com/charmbracelet/bubbletea"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type Source struct {
	columns   []*source.List
	clientset *kubernetes.Clientset
	cfg       *config.Config
}

func getClientset(k8sContext string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	if home := homedir.HomeDir(); home != "" {
		kubeconfig := filepath.Join(home, ".kube", "config")
		// Use the provided K8sContext if it's not empty
		if k8sContext != "" {
			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			loadingRules.ExplicitPath = kubeconfig
			configOverrides := &clientcmd.ConfigOverrides{CurrentContext: k8sContext}
			kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
			config, err = kubeConfig.ClientConfig()
		} else {
			config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		}
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("Home directory not found, unable to locate kubeconfig file")
	}

	config.QPS = 40
	config.Burst = 40
	config.WarningHandler = rest.NoWarnings{}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func New(config *config.Config) source.Source {
	src := &Source{
		columns: []*source.List{
			source.NewList("Namespaces", []source.ListItem{}),
			source.NewList("Workloads", []source.ListItem{}),
			source.NewList("Pods", []source.ListItem{}),
			source.NewList("Containers", []source.ListItem{}),
		},
		cfg: config,
	}
	return src
}

func (s *Source) Init(stateChan chan state.State) tea.Cmd {
	return func() tea.Msg {
		stateChan <- state.StateLoading

		clientset, err := getClientset(s.cfg.K8sContext)
		if err != nil {
			return err
		}

		s.clientset = clientset

		namespaces := make([]*Namespace, len(s.cfg.Namespaces))
		var wg sync.WaitGroup
		wg.Add(len(s.cfg.Namespaces))
		for i, namespace := range s.cfg.Namespaces {
			go func(i int, namespace string) {
				n := NewNamespace(namespace, clientset)
				namespaces[i] = n
				wg.Done()
			}(i, namespace)
		}
		wg.Wait()

		s.columns[0].SetItems(source.ConvertInterface2ListItems(namespaces))
		s.columns[0].ResetFilter()
		s.columns[0].ResetSelected()

		for i := 0; i < len(s.columns)-1; i++ {
			if item := s.columns[i].SelectedItem(); item != nil {
				s.columns[i+1].SetItems(item.Children())
			} else {
				s.columns[i+1].SetItems([]source.ListItem{})
			}
			s.columns[i+1].ResetFilter()
			s.columns[i+1].ResetSelected()
		}

		stateChan <- state.StateBrose
		return nil
	}
}

func (s *Source) Columns() []*source.List {
	return s.columns
}

func (s *Source) getLogCfg() map[string]string {
	cfg := map[string]string{}
	for _, c := range s.columns {
		s := c.SelectedItem()
		cfg[strings.ToLower(c.Title)] = s.String()
		if c.IsActive() {
			break
		}
	}
	return cfg
}

func (s *Source) Logs(ctx context.Context, stateChan chan state.State, logChan chan string) tea.Cmd {
	cfg := s.getLogCfg()
	return func() tea.Msg {

		if c, ok := cfg["containers"]; c == "" || !ok {
			stateChan <- state.StateBrose
			return nil
		}

		tail := int64(10000)
		podLogOpts := v1.PodLogOptions{
			Container: cfg["containers"],
			TailLines: &tail,
			Follow:    true,
		}

		req := s.clientset.CoreV1().Pods(cfg["namespaces"]).GetLogs(cfg["pods"], &podLogOpts)

		podLogs, err := req.Stream(ctx)
		if err != nil {
			return err
		}
		defer podLogs.Close()

		scanner := bufio.NewScanner(podLogs)
		const maxCapacity = 1024 * 1024 // 1MB, default was 64kb
		buf := make([]byte, 0, maxCapacity)
		scanner.Buffer(buf, maxCapacity)
		stateChan <- state.StateLogs

		if scanner.Scan() {
			stateChan <- state.StateLogs
			logChan <- scanner.Text()
		}

		for scanner.Scan() {
			t := scanner.Text()
			logChan <- t
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		return nil
	}
}

func (s *Source) Close() {
}
