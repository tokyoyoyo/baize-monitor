package network

import (
	"os/exec"
	"strings"

	"baize-monitor/internal/agent/anomaly/utils"
	"baize-monitor/pkg/dto/request"
	"baize-monitor/pkg/models"
)

type networkLinkStatusDetector struct {
}

func init() {
	RegisterDetector(&networkLinkStatusDetector{})
}

func (d *networkLinkStatusDetector) CheckItemName() string {
	return "network_link_down"
}

func (d *networkLinkStatusDetector) Detect() (request.AnomalyResult, error) {
	cmd := exec.Command("ip", "-o", "link")
	output, err := cmd.Output()
	if err != nil {
		return request.AnomalyResult{}, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		if strings.Contains(line, "state DOWN") &&
			!strings.Contains(line, "lo:") {
			parts := strings.Fields(line)
			iface := ""
			for i, part := range parts {
				if strings.HasSuffix(part, ":") {
					iface = strings.TrimSuffix(part, ":")
					break
				}
				if i > 2 {
					break
				}
			}

			if iface != "" {
				return utils.CreateAnomalyResult(
					models.CheckTypeNetwork,
					d.CheckItemName(),
					models.AnomalyLevelCritical,
					"Network interface link is down",
					map[string]interface{}{
						"interface": iface,
						"status":    "DOWN",
						"issue":     "link_down",
					},
				), nil
			}
		}
	}

	return request.AnomalyResult{}, nil
}
