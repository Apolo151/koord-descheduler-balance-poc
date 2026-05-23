package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"koord-descheduler-balance-poc/pkg/binpack"
	"koord-descheduler-balance-poc/pkg/descheduler"
	"koord-descheduler-balance-poc/pkg/fixtures"
	"koord-descheduler-balance-poc/pkg/imbalance"
	"koord-descheduler-balance-poc/pkg/model"
	"koord-descheduler-balance-poc/pkg/report"
)

func main() {
	scenarioPath := flag.String("scenario", "", "path to scenario YAML/JSON")
	outJSON := flag.String("out-json", "", "output JSON file (default: skip)")
	outSummary := flag.String("out-summary", "", "output summary file (default: stdout)")
	outHTML := flag.String("out-html", "", "output HTML file (default: skip)")
	outHTMLDir := flag.String("out-html-dir", "", "output HTML directory (default: skip)")
	serveHTML := flag.Bool("serve-html", false, "serve HTML output over HTTP")
	serveAddr := flag.String("serve-addr", "127.0.0.1:8080", "address to serve HTML (default: 127.0.0.1:8080)")
	flag.Parse()

	if *scenarioPath == "" {
		fmt.Fprintln(os.Stderr, "--scenario is required")
		os.Exit(2)
	}

	scenario, err := fixtures.LoadScenario(*scenarioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load scenario: %v\n", err)
		os.Exit(1)
	}
	if err := fixtures.ValidateScenario(scenario); err != nil {
		fmt.Fprintf(os.Stderr, "validate scenario: %v\n", err)
		os.Exit(1)
	}

	snapshot, imbalanceCfg, evictionPolicy, binpackCfg := fixtures.ToSnapshot(scenario)
	result := report.Result{
		Scores:  computeScores(snapshot, imbalanceCfg),
		Plans:   descheduler.PlanEvictions(snapshot, imbalanceCfg, evictionPolicy),
		Binpack: nil,
	}

	if binpackCfg.Victims > 0 && (binpackCfg.TargetWorkload != "" || len(binpackCfg.Selector) > 0) {
		decision := binpack.SelectVictims(snapshot, binpackCfg)
		result.Binpack = &decision
	}

	if *outJSON != "" {
		if err := writeJSON(*outJSON, result); err != nil {
			fmt.Fprintf(os.Stderr, "write json: %v\n", err)
			os.Exit(1)
		}
	}

	htmlPath, err := writeHTML(*outHTML, *outHTMLDir, *scenarioPath, scenario.Metadata, snapshot, result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write html: %v\n", err)
		os.Exit(1)
	}

	summary := report.RenderSummary(result)
	if *outSummary != "" {
		if err := os.WriteFile(*outSummary, []byte(summary), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "write summary: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Print(summary)
	}

	if *serveHTML {
		if htmlPath == "" {
			fmt.Fprintln(os.Stderr, "serve html: no HTML output path; use --out-html or --out-html-dir")
			os.Exit(1)
		}
		if err := serveHTMLFile(htmlPath, *serveAddr); err != nil {
			fmt.Fprintf(os.Stderr, "serve html: %v\n", err)
			os.Exit(1)
		}
	}
}

func writeJSON(path string, result report.Result) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func writeHTML(outHTML string, outHTMLDir string, scenarioPath string, metadata map[string]string, snapshot *model.ClusterSnapshot, result report.Result) (string, error) {
	outputPath, err := resolveHTMLOutput(outHTML, outHTMLDir, scenarioPath, metadata)
	if err != nil {
		return "", err
	}
	if outputPath == "" {
		return "", nil
	}

	title := scenarioTitle(scenarioPath, metadata)
	subtitle := "Eviction and binpack visualization"
	html, err := report.RenderHTML(snapshot, result, report.HTMLMetadata{
		Title:       title,
		Subtitle:    subtitle,
		ScenarioRef: filepath.Base(scenarioPath),
	})
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return "", err
	}
	if err := os.WriteFile(outputPath, []byte(html), 0644); err != nil {
		return "", err
	}
	return outputPath, nil
}

func serveHTMLFile(path string, addr string) error {
	if strings.TrimSpace(addr) == "" {
		return fmt.Errorf("serve address is empty")
	}
	base := filepath.Base(path)
	root := filepath.Dir(path)
	url := fmt.Sprintf("http://%s/%s", addr, base)
	fmt.Printf("Serving %s\n", url)
	return http.ListenAndServe(addr, http.FileServer(http.Dir(root)))
}

func resolveHTMLOutput(outHTML string, outHTMLDir string, scenarioPath string, metadata map[string]string) (string, error) {
	if strings.TrimSpace(outHTML) != "" {
		return outHTML, nil
	}
	if strings.TrimSpace(outHTMLDir) == "" {
		return "", nil
	}

	base := scenarioTitle(scenarioPath, metadata)
	if base == "" {
		base = strings.TrimSuffix(filepath.Base(scenarioPath), filepath.Ext(scenarioPath))
	}
	base = sanitizeFilename(base)
	if base == "" {
		base = "scenario"
	}
	return filepath.Join(outHTMLDir, base+".html"), nil
}

func scenarioTitle(scenarioPath string, metadata map[string]string) string {
	if metadata != nil {
		if name := strings.TrimSpace(metadata["name"]); name != "" {
			return name
		}
	}
	return strings.TrimSuffix(filepath.Base(scenarioPath), filepath.Ext(scenarioPath))
}

func sanitizeFilename(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var builder strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			builder.WriteRune(r)
			continue
		}
		if r == ' ' || r == '.' {
			builder.WriteRune('-')
		}
	}
	return strings.Trim(builder.String(), "-")
}

func computeScores(snapshot *model.ClusterSnapshot, cfg model.ImbalanceConfig) []model.NodeImbalanceScore {
	if snapshot == nil || snapshot.Nodes == nil {
		return nil
	}

	nodeNames := make([]string, 0, len(snapshot.Nodes))
	for name := range snapshot.Nodes {
		nodeNames = append(nodeNames, name)
	}
	sort.Strings(nodeNames)

	scores := make([]model.NodeImbalanceScore, 0, len(nodeNames))
	for _, nodeName := range nodeNames {
		node := snapshot.Nodes[nodeName]
		if node == nil {
			continue
		}
		scores = append(scores, imbalance.ScoreNode(node, snapshot, cfg))
	}
	return scores
}
