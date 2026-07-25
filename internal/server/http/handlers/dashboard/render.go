package dashboard

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"
	"strings"
	"time"

	"consul-journey/internal"
	"consul-journey/internal/errors"
	"consul-journey/internal/node"

	"github.com/gofiber/fiber/v3"
)

var tmpl = template.Must(
	template.New("dashboard").Funcs(funcMap).Parse(dashboardTemplates),
)

var funcMap = template.FuncMap{
	"statusClass": statusClass,
	"statusLabel": statusLabel,
	"httpAddr": func(p *node.Peer) string {
		return fmt.Sprintf("%s:%d", p.Address, p.HTTPPort)
	},
}

type baseView struct {
	AppName     string
	Version     string
	Revision    string
	GeneratedAt string
}

type peerRow struct {
	*node.Peer
	IsCurrent bool
	IsLeader  bool
}

type kv struct {
	Key   string
	Value string
}

type peersView struct {
	baseView
	Rows         []peerRow
	Total        int
	Healthy      int
	ElectionOn   bool
	LeaderKnown  bool
	LeaderNode   string
	LeaderID     string
	PeerBasePath string
}

type peerView struct {
	baseView
	Row        peerRow
	Meta       []kv
	ElectionOn bool
	ListPath   string
}

type errorView struct {
	baseView
	Status  int
	Code    string
	Message string
	Path    string
	Hint    string
}

func (h *Handler) base() baseView {
	return baseView{
		AppName:     internal.AppName(),
		Version:     internal.Version(),
		Revision:    internal.Revision(),
		GeneratedAt: time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
	}
}

func (h *Handler) render(c fiber.Ctx, status int, name string, data any) error {
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return errors.NewInternal(c.Method()+" "+c.OriginalURL(), err)
	}
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return c.Status(status).Send(buf.Bytes())
}

func (h *Handler) renderError(c fiber.Ctx, e *errors.Error) error {
	view := errorView{
		baseView: h.base(),
		Status:   e.Status(),
		Code:     e.Code(),
		Message:  e.Error(),
		Path:     c.Method() + " " + c.OriginalURL(),
		Hint:     errorHint(e.Status()),
	}
	return h.render(c, e.Status(), "error", view)
}

func statusClass(s string) string {
	switch strings.ToLower(s) {
	case "passing":
		return "passing"
	case "warning":
		return "warning"
	case "critical":
		return "critical"
	case "maintenance":
		return "maintenance"
	default:
		return "unknown"
	}
}

func statusLabel(s string) string {
	if s == "" {
		return "Unknown"
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

func errorHint(status int) string {
	switch status {
	case fiber.StatusNotFound:
		return "The peer may have left the cluster, or the identifier is incorrect."
	case fiber.StatusBadRequest:
		return "The request was malformed. Check the address bar and try again."
	default:
		return "Something went wrong while serving this page."
	}
}

func sortedMeta(m map[string]string) []kv {
	out := make([]kv, 0, len(m))
	for k, v := range m {
		out = append(out, kv{Key: k, Value: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

func isPassing(status string) bool {
	return strings.EqualFold(status, "passing")
}
