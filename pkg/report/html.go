package report

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"sort"
	"strings"

	"koord-descheduler-balance-poc/pkg/model"
)

type HTMLMetadata struct {
	Title       string
	Subtitle    string
	ScenarioRef string
}

type htmlReport struct {
	Title         string
	Subtitle      string
	ScenarioRef   string
	NodeCount     int
	EvictionCount int
	VictimCount   int
	Nodes         []htmlNode
	EvictedPods   []htmlPod
	HasBinpack    bool
	Victims       []string
	Remaining     []string
}

type htmlNode struct {
	Name           string
	Mean           string
	StdDev         string
	CandidateCount int
	ChosenCount    int
	IsVictim       bool
	Pods           []htmlPod
}

type htmlPod struct {
	ID        string
	Name      string
	Owner     string
	Namespace string
	Evictable bool
	Evicted   bool
	Resources []htmlResource
}

type htmlResource struct {
	Name  string
	Value string
}

func RenderHTML(snapshot *model.ClusterSnapshot, result Result, meta HTMLMetadata) (string, error) {
	if snapshot == nil {
		return "", errors.New("snapshot is nil")
	}

	chosenPods := map[string]bool{}
	planCounts := map[string]struct {
		candidates int
		chosen     int
	}{}
	for _, plan := range result.Plans {
		planCounts[plan.NodeName] = struct {
			candidates int
			chosen     int
		}{
			candidates: len(plan.Candidates),
			chosen:     len(plan.Chosen),
		}
		for _, pod := range plan.Chosen {
			chosenPods[pod.PodID] = true
		}
	}

	victimNodes := map[string]bool{}
	if result.Binpack != nil {
		for _, node := range result.Binpack.Victims {
			victimNodes[node] = true
		}
	}

	scoreByNode := map[string]model.NodeImbalanceScore{}
	for _, score := range result.Scores {
		scoreByNode[score.NodeName] = score
	}

	nodeNames := make([]string, 0, len(snapshot.Nodes))
	for name := range snapshot.Nodes {
		nodeNames = append(nodeNames, name)
	}
	sort.Strings(nodeNames)

	report := htmlReport{
		Title:         meta.Title,
		Subtitle:      meta.Subtitle,
		ScenarioRef:   meta.ScenarioRef,
		NodeCount:     len(nodeNames),
		EvictionCount: len(chosenPods),
		VictimCount:   len(victimNodes),
		HasBinpack:    result.Binpack != nil,
	}
	if result.Binpack != nil {
		report.Victims = append([]string{}, result.Binpack.Victims...)
		report.Remaining = append([]string{}, result.Binpack.NodeSet...)
		sort.Strings(report.Victims)
		sort.Strings(report.Remaining)
	}

	report.Nodes = make([]htmlNode, 0, len(nodeNames))
	for _, nodeName := range nodeNames {
		node := snapshot.Nodes[nodeName]
		if node == nil {
			continue
		}
		score := scoreByNode[nodeName]
		counts := planCounts[nodeName]

		hNode := htmlNode{
			Name:           nodeName,
			Mean:           formatFloat(score.Mean),
			StdDev:         formatFloat(score.StdDev),
			CandidateCount: counts.candidates,
			ChosenCount:    counts.chosen,
			IsVictim:       victimNodes[nodeName],
		}

		pods := make([]htmlPod, 0, len(node.Pods))
		for _, podID := range node.Pods {
			pod := snapshot.Pods[podID]
			if pod == nil {
				continue
			}
			podView := htmlPod{
				ID:        podID,
				Name:      pod.Name,
				Owner:     emptyFallback(pod.Owner, "none"),
				Namespace: emptyFallback(pod.Namespace, "default"),
				Evictable: pod.Evictable,
				Evicted:   chosenPods[podID],
			}
			resourceKeys := pod.Resources.Keys()
			podView.Resources = make([]htmlResource, 0, len(resourceKeys))
			for _, key := range resourceKeys {
				podView.Resources = append(podView.Resources, htmlResource{
					Name:  key,
					Value: formatFloat(pod.Resources.Get(key)),
				})
			}
			pods = append(pods, podView)
			if podView.Evicted {
				report.EvictedPods = append(report.EvictedPods, podView)
			}
		}

		hNode.Pods = pods
		report.Nodes = append(report.Nodes, hNode)
	}

	sort.SliceStable(report.EvictedPods, func(i, j int) bool {
		return report.EvictedPods[i].ID < report.EvictedPods[j].ID
	})

	var buffer bytes.Buffer
	if err := htmlTemplate.Execute(&buffer, report); err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func emptyFallback(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func formatFloat(value float64) string {
	formatted := fmt.Sprintf("%.3f", value)
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimRight(formatted, ".")
	if formatted == "" {
		return "0"
	}
	return formatted
}

var htmlTemplate = template.Must(template.New("report").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>{{ .Title }}</title>
  <style>
    :root {
      --bg: #0f172a;
      --bg-soft: #111c36;
      --panel: #f8fafc;
      --panel-soft: #eef2f7;
      --ink: #0f172a;
      --ink-soft: #1f2937;
      --accent: #f97316;
      --accent-2: #14b8a6;
      --danger: #ef4444;
      --muted: #64748b;
    }

    * { box-sizing: border-box; }

    body {
      margin: 0;
      font-family: "Gill Sans", "Trebuchet MS", "Calibri", sans-serif;
      color: var(--ink);
      background-image:
        radial-gradient(circle at 20% 20%, rgba(249, 115, 22, 0.15), transparent 45%),
        radial-gradient(circle at 80% 10%, rgba(20, 184, 166, 0.18), transparent 40%),
        linear-gradient(140deg, #0b1120 0%, #111827 45%, #0f172a 100%);
      min-height: 100vh;
      padding: 32px;
    }

    .page {
      max-width: 1200px;
      margin: 0 auto;
      background: var(--panel);
      border-radius: 28px;
      box-shadow: 0 20px 60px rgba(15, 23, 42, 0.35);
      padding: 32px;
    }

    .hero {
      display: flex;
      gap: 24px;
      justify-content: space-between;
      align-items: flex-start;
      flex-wrap: wrap;
      margin-bottom: 24px;
    }

    .kicker {
      font-size: 12px;
      letter-spacing: 2px;
      text-transform: uppercase;
      color: var(--muted);
      font-weight: 700;
    }

    h1 {
      font-family: "Palatino Linotype", "Book Antiqua", Palatino, serif;
      margin: 8px 0 6px;
      font-size: 32px;
      color: var(--ink);
    }

    .subtitle {
      margin: 0;
      color: var(--ink-soft);
      max-width: 520px;
    }

    .meta {
      display: grid;
      grid-template-columns: repeat(3, minmax(120px, 1fr));
      gap: 12px;
      min-width: 280px;
    }

    .meta-card {
      background: var(--panel-soft);
      border-radius: 14px;
      padding: 12px 14px;
      border: 1px solid rgba(15, 23, 42, 0.08);
    }

    .meta-label {
      font-size: 12px;
      text-transform: uppercase;
      letter-spacing: 1px;
      color: var(--muted);
      margin-bottom: 4px;
    }

    .meta-value {
      font-size: 20px;
      font-weight: 700;
    }

    .controls {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 12px;
      margin-bottom: 22px;
      padding: 14px;
      border-radius: 18px;
      background: var(--panel-soft);
      border: 1px solid rgba(15, 23, 42, 0.08);
    }

    .control-group {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
      align-items: center;
    }

    .control-label {
      font-size: 12px;
      text-transform: uppercase;
      letter-spacing: 1px;
      color: var(--muted);
      font-weight: 700;
      margin-right: 6px;
    }

    .btn {
      border: 1px solid rgba(15, 23, 42, 0.18);
      background: #fff;
      color: var(--ink);
      padding: 8px 14px;
      border-radius: 999px;
      cursor: pointer;
      font-weight: 600;
      transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
    }

    .btn:active {
      transform: translateY(1px);
    }

    .btn-primary {
      border-color: rgba(249, 115, 22, 0.7);
      color: #fff;
      background: linear-gradient(120deg, #f97316, #f59e0b);
      box-shadow: 0 10px 22px rgba(249, 115, 22, 0.28);
    }

    .btn-secondary {
      background: #fff;
    }

    .btn.active {
      border-color: var(--accent-2);
      color: var(--accent-2);
      box-shadow: 0 10px 20px rgba(20, 184, 166, 0.2);
    }

    .progress {
      position: relative;
      height: 12px;
      background: rgba(15, 23, 42, 0.12);
      border-radius: 999px;
      overflow: hidden;
      flex: 1 1 auto;
      min-width: 160px;
    }

    .progress-fill {
      height: 100%;
      width: 0%;
      background: linear-gradient(90deg, var(--accent), var(--accent-2));
      transition: width 0.35s ease;
    }

    .progress-label {
      font-size: 12px;
      color: var(--muted);
      font-weight: 600;
      margin-left: 8px;
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
      gap: 16px;
    }

    .node {
      border-radius: 20px;
      background: #fff;
      border: 1px solid rgba(15, 23, 42, 0.08);
      padding: 16px;
      display: flex;
      flex-direction: column;
      gap: 12px;
      transition: transform 0.3s ease, box-shadow 0.3s ease, opacity 0.3s ease;
    }

    .node-head {
      display: flex;
      flex-direction: column;
      gap: 4px;
    }

    .node-name {
      font-weight: 700;
      font-size: 18px;
    }

    .node-score,
    .node-meta {
      font-size: 13px;
      color: var(--muted);
    }

    .pod-list {
      display: flex;
      flex-direction: column;
      gap: 10px;
    }

    .pod {
      border-radius: 16px;
      padding: 10px 12px;
      background: linear-gradient(120deg, rgba(249, 115, 22, 0.1), rgba(20, 184, 166, 0.12));
      border: 1px solid rgba(15, 23, 42, 0.08);
      display: flex;
      flex-direction: column;
      gap: 6px;
      transition: transform 0.4s ease, opacity 0.4s ease, box-shadow 0.4s ease;
    }

    .pod[data-evictable="false"] {
      background: #f1f5f9;
      color: var(--muted);
    }

    .pod-title {
      display: flex;
      justify-content: space-between;
      gap: 6px;
      font-weight: 600;
      font-size: 14px;
    }

    .pod-id {
      font-size: 12px;
      color: var(--muted);
    }

    .pod-resources {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
    }

    .chip {
      font-size: 11px;
      padding: 3px 8px;
      border-radius: 999px;
      background: #fff;
      border: 1px solid rgba(15, 23, 42, 0.08);
    }

    .pulse {
      animation: pulse 1.4s ease-in-out infinite;
    }

    .tray {
      margin-top: 24px;
      padding: 16px;
      border-radius: 20px;
      background: #0f172a;
      color: #f8fafc;
    }

    .tray-head {
      font-size: 13px;
      text-transform: uppercase;
      letter-spacing: 1px;
      margin-bottom: 12px;
      color: #94a3b8;
    }

    .tray-list {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
    }

    .tray .pod {
      background: rgba(248, 250, 252, 0.12);
      border-color: rgba(248, 250, 252, 0.2);
      color: #f8fafc;
    }

    .tray .chip {
      background: rgba(15, 23, 42, 0.6);
      color: #f8fafc;
    }

    .empty {
      font-size: 12px;
      color: #94a3b8;
      padding: 6px 0;
    }

    .binpack {
      margin-top: 20px;
      padding: 14px 16px;
      border-radius: 16px;
      background: var(--panel-soft);
      border: 1px solid rgba(15, 23, 42, 0.08);
    }

    .binpack h3 {
      margin: 0 0 10px;
      font-size: 14px;
      text-transform: uppercase;
      letter-spacing: 1px;
      color: var(--muted);
    }

    .binpack-row {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      font-size: 13px;
      margin-bottom: 8px;
    }

    .tag {
      padding: 4px 10px;
      border-radius: 999px;
      background: #fff;
      border: 1px solid rgba(15, 23, 42, 0.12);
      font-weight: 600;
    }

    body[data-phase="1"] .pod[data-evict="true"] {
      opacity: 0.15;
      transform: translateY(-14px) scale(0.98);
      box-shadow: 0 16px 26px rgba(239, 68, 68, 0.12);
    }

    body[data-phase="1"] .tray {
      box-shadow: 0 18px 32px rgba(15, 23, 42, 0.45);
    }

    body[data-phase="2"] .node.is-victim {
      opacity: 0.35;
      transform: scale(0.97);
      filter: grayscale(0.5);
    }

    body[data-phase="2"] .node:not(.is-victim) {
      box-shadow: 0 14px 24px rgba(20, 184, 166, 0.2);
      border-color: rgba(20, 184, 166, 0.45);
    }

    body[data-phase="2"] .node:not(.is-victim) .node-name {
      color: var(--accent-2);
    }

    @keyframes pulse {
      0% { transform: translateY(0) scale(1); }
      50% { transform: translateY(-2px) scale(1.01); }
      100% { transform: translateY(0) scale(1); }
    }

    @media (max-width: 720px) {
      body { padding: 16px; }
      .page { padding: 20px; }
      .meta { grid-template-columns: repeat(2, minmax(120px, 1fr)); }
    }
  </style>
</head>
<body data-phase="0">
  <div class="page">
    <header class="hero">
      <div>
        <div class="kicker">Koord Balance Poc</div>
        <h1>{{ .Title }}</h1>
        <p class="subtitle">{{ .Subtitle }}</p>
        {{ if .ScenarioRef }}
          <p class="subtitle">Scenario: {{ .ScenarioRef }}</p>
        {{ end }}
      </div>
      <div class="meta">
        <div class="meta-card">
          <div class="meta-label">Nodes</div>
          <div class="meta-value">{{ .NodeCount }}</div>
        </div>
        <div class="meta-card">
          <div class="meta-label">Evictions</div>
          <div class="meta-value">{{ .EvictionCount }}</div>
        </div>
        <div class="meta-card">
          <div class="meta-label">Binpack Victims</div>
          <div class="meta-value">{{ .VictimCount }}</div>
        </div>
      </div>
    </header>

    <section class="controls">
      <div class="control-group">
        <span class="control-label">Run</span>
        <button class="btn btn-primary" type="button" id="runAll">Run All</button>
        <button class="btn btn-secondary" type="button" id="togglePause">Pause</button>
        <button class="btn btn-secondary" type="button" id="resetRun">Reset</button>
      </div>
      <div class="control-group">
        <span class="control-label">Phase</span>
        <button class="btn" type="button" data-phase-btn="0">Initial</button>
        <button class="btn" type="button" data-phase-btn="1">Evictions</button>
        <button class="btn" type="button" data-phase-btn="2">Binpack</button>
      </div>
      <div class="control-group">
        <span class="control-label">Progress</span>
        <div class="progress"><div class="progress-fill" id="progressFill"></div></div>
        <span class="progress-label" id="progressLabel">0%</span>
      </div>
    </section>

    <section class="grid">
      {{ range .Nodes }}
      <article class="node {{ if .IsVictim }}is-victim{{ end }}">
        <div class="node-head">
          <div class="node-name">{{ .Name }}</div>
          <div class="node-score">mean {{ .Mean }} | std {{ .StdDev }}</div>
          <div class="node-meta">candidates {{ .CandidateCount }} | chosen {{ .ChosenCount }}</div>
        </div>
        <div class="pod-list">
          {{ range .Pods }}
          <div class="pod" data-evict="{{ .Evicted }}" data-evictable="{{ .Evictable }}">
            <div class="pod-title">
              <span>{{ .Name }}</span>
              <span>{{ .Owner }}</span>
            </div>
            <div class="pod-id">{{ .Namespace }}/{{ .Name }}</div>
            <div class="pod-resources">
              {{ range .Resources }}
                <span class="chip">{{ .Name }}: {{ .Value }}</span>
              {{ end }}
            </div>
          </div>
          {{ end }}
        </div>
      </article>
      {{ end }}
    </section>

    <section class="tray">
      <div class="tray-head">Evicted Pods</div>
      <div class="tray-list">
        {{ if .EvictedPods }}
          {{ range .EvictedPods }}
            <div class="pod" data-evict="true">
              <div class="pod-title">
                <span>{{ .Name }}</span>
                <span>{{ .Owner }}</span>
              </div>
              <div class="pod-id">{{ .ID }}</div>
              <div class="pod-resources">
                {{ range .Resources }}
                  <span class="chip">{{ .Name }}: {{ .Value }}</span>
                {{ end }}
              </div>
            </div>
          {{ end }}
        {{ else }}
          <div class="empty">No evictions chosen in this plan.</div>
        {{ end }}
      </div>
    </section>

    <section class="binpack">
      <h3>Binpack Decision</h3>
      {{ if .HasBinpack }}
        <div class="binpack-row">
          <strong>Victims:</strong>
          {{ range .Victims }}
            <span class="tag">{{ . }}</span>
          {{ end }}
        </div>
        <div class="binpack-row">
          <strong>Nodes Remaining:</strong>
          {{ range .Remaining }}
            <span class="tag">{{ . }}</span>
          {{ end }}
        </div>
      {{ else }}
        <div class="empty">No binpack decision in this scenario.</div>
      {{ end }}
    </section>
  </div>

  <script>
    (function () {
      var buttons = document.querySelectorAll('[data-phase-btn]');
      var runButton = document.getElementById('runAll');
      var pauseButton = document.getElementById('togglePause');
      var resetButton = document.getElementById('resetRun');
      var progressFill = document.getElementById('progressFill');
      var progressLabel = document.getElementById('progressLabel');

      var phaseOrder = ['0', '1', '2'];
      var phaseDuration = 1400;
      var progressTick = 60;
      var currentPhaseIndex = 0;
      var isRunning = false;
      var isPaused = false;
      var runTimer = null;
      var progressTimer = null;

      function setPhase(value) {
        document.body.setAttribute('data-phase', value);
        buttons.forEach(function (button) {
          button.classList.toggle('active', button.getAttribute('data-phase-btn') === value);
        });
      }

      function updateProgress(percent) {
        progressFill.style.width = percent + '%';
        progressLabel.textContent = Math.round(percent) + '%';
      }

      function stopTimers() {
        if (runTimer) {
          clearTimeout(runTimer);
          runTimer = null;
        }
        if (progressTimer) {
          clearInterval(progressTimer);
          progressTimer = null;
        }
      }

      function setButtonStates() {
        pauseButton.textContent = isPaused ? 'Resume' : 'Pause';
        pauseButton.disabled = !isRunning;
      }

      function stepPhase() {
        if (!isRunning || isPaused) {
          return;
        }
        setPhase(phaseOrder[currentPhaseIndex]);
        animateProgress();
        currentPhaseIndex += 1;

        if (currentPhaseIndex >= phaseOrder.length) {
          runTimer = setTimeout(function () {
            isRunning = false;
            isPaused = false;
            setButtonStates();
          }, phaseDuration);
          return;
        }

        runTimer = setTimeout(stepPhase, phaseDuration);
      }

      function animateProgress() {
        var start = currentPhaseIndex / phaseOrder.length * 100;
        var end = (currentPhaseIndex + 1) / phaseOrder.length * 100;
        var elapsed = 0;
        updateProgress(start);
        if (progressTimer) {
          clearInterval(progressTimer);
        }
        progressTimer = setInterval(function () {
          if (isPaused || !isRunning) {
            clearInterval(progressTimer);
            progressTimer = null;
            return;
          }
          elapsed += progressTick;
          var pct = Math.min(end, start + (elapsed / phaseDuration) * (end - start));
          updateProgress(pct);
          if (pct >= end) {
            clearInterval(progressTimer);
            progressTimer = null;
          }
        }, progressTick);
      }

      function runAll() {
        stopTimers();
        isRunning = true;
        isPaused = false;
        currentPhaseIndex = 0;
        setButtonStates();
        stepPhase();
      }

      function pauseOrResume() {
        if (!isRunning) {
          return;
        }
        isPaused = !isPaused;
        setButtonStates();
        if (!isPaused) {
          stepPhase();
        }
      }

      function resetRun() {
        stopTimers();
        isRunning = false;
        isPaused = false;
        currentPhaseIndex = 0;
        setPhase('0');
        updateProgress(0);
        setButtonStates();
      }

      buttons.forEach(function (button) {
        button.addEventListener('click', function () {
          stopTimers();
          isRunning = false;
          isPaused = false;
          currentPhaseIndex = phaseOrder.indexOf(button.getAttribute('data-phase-btn'));
          setPhase(button.getAttribute('data-phase-btn'));
          updateProgress((currentPhaseIndex + 1) / phaseOrder.length * 100);
          setButtonStates();
        });
      });

      runButton.addEventListener('click', runAll);
      pauseButton.addEventListener('click', pauseOrResume);
      resetButton.addEventListener('click', resetRun);

      resetRun();
    })();
  </script>
</body>
</html>`))
