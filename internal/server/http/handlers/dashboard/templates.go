package dashboard

const dashboardTemplates = `
{{define "styles"}}
<style>
  :root {
    --bg: #0f172a;
    --bg-grad-1: #1e293b;
    --bg-grad-2: #0f172a;
    --surface: #ffffff;
    --surface-2: #f8fafc;
    --border: #e2e8f0;
    --text: #0f172a;
    --text-muted: #64748b;
    --text-faint: #94a3b8;
    --brand: #4f46e5;
    --brand-soft: #eef2ff;
    --accent: #6366f1;
    --current: #eff6ff;
    --current-bar: #3b82f6;
    --leader: #b45309;
    --leader-soft: #fffbeb;
    --leader-border: #fcd34d;
    --ok: #16a34a;
    --ok-soft: #f0fdf4;
    --ok-border: #bbf7d0;
    --warn: #d97706;
    --warn-soft: #fffbeb;
    --warn-border: #fde68a;
    --crit: #dc2626;
    --crit-soft: #fef2f2;
    --crit-border: #fecaca;
    --muted-soft: #f1f5f9;
    --muted-border: #e2e8f0;
    --shadow: 0 1px 2px rgba(15,23,42,.04), 0 8px 24px rgba(15,23,42,.06);
    --radius: 14px;
    --mono: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
    --sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  }
  @media (prefers-color-scheme: dark) {
    :root {
      --surface: #111827;
      --surface-2: #0b1220;
      --border: #1f2937;
      --text: #e5e7eb;
      --text-muted: #9ca3af;
      --text-faint: #6b7280;
      --brand-soft: #1e1b4b;
      --current: #0b2545;
      --leader-soft: #2a1e05;
      --leader-border: #7c5b12;
      --ok-soft: #052e1a;
      --ok-border: #14532d;
      --warn-soft: #2a1e05;
      --warn-border: #7c5b12;
      --crit-soft: #2a0d0d;
      --crit-border: #7f1d1d;
      --muted-soft: #1f2937;
      --muted-border: #374151;
      --shadow: 0 1px 2px rgba(0,0,0,.3), 0 12px 32px rgba(0,0,0,.35);
    }
  }
  * { box-sizing: border-box; }
  html, body { margin: 0; padding: 0; }
  body {
    font-family: var(--sans);
    color: var(--text);
    background: var(--surface-2);
    -webkit-font-smoothing: antialiased;
    line-height: 1.5;
  }
  a { color: var(--brand); text-decoration: none; }
  .wrap { max-width: 1120px; margin: 0 auto; padding: 0 24px 64px; }

  /* Header */
  .masthead {
    background: linear-gradient(135deg, var(--bg-grad-1), var(--bg-grad-2));
    color: #e2e8f0;
    padding: 28px 0 30px;
    border-bottom: 1px solid rgba(148,163,184,.12);
  }
  .masthead .wrap { padding-bottom: 0; }
  .brandline { display: flex; align-items: center; gap: 12px; }
  .logo {
    width: 38px; height: 38px; border-radius: 10px; flex: none;
    background: linear-gradient(135deg, var(--brand), #a855f7);
    display: grid; place-items: center; color: #fff; font-weight: 700;
    box-shadow: 0 6px 16px rgba(79,70,229,.4);
  }
  .brandline h1 { font-size: 19px; margin: 0; font-weight: 650; letter-spacing: -.01em; color: #fff; }
  .brandline .sub { font-size: 12.5px; color: #94a3b8; margin-top: 1px; }
  .crumbs { margin-top: 4px; font-size: 12.5px; color: #94a3b8; }
  .crumbs a { color: #c7d2fe; }

  /* Stat chips */
  .stats { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 22px; }
  .stat {
    background: rgba(255,255,255,.06);
    border: 1px solid rgba(148,163,184,.18);
    border-radius: 12px; padding: 12px 16px; min-width: 128px;
    backdrop-filter: blur(6px);
  }
  .stat .label { font-size: 11px; text-transform: uppercase; letter-spacing: .06em; color: #94a3b8; }
  .stat .value { font-size: 20px; font-weight: 680; color: #fff; margin-top: 3px; display: flex; align-items: center; gap: 8px; }
  .stat .value.small { font-size: 15px; font-weight: 600; }

  /* Card */
  .card {
    background: var(--surface);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    box-shadow: var(--shadow);
    margin-top: 26px;
    overflow: hidden;
  }
  .card-head {
    display: flex; align-items: center; justify-content: space-between; gap: 12px;
    padding: 16px 20px; border-bottom: 1px solid var(--border);
  }
  .card-head h2 { font-size: 14px; margin: 0; font-weight: 640; letter-spacing: -.01em; }
  .card-head .meta { font-size: 12px; color: var(--text-muted); }

  /* Table */
  .tbl-scroll { width: 100%; overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; font-size: 13.5px; }
  thead th {
    text-align: left; font-size: 11px; text-transform: uppercase; letter-spacing: .05em;
    color: var(--text-muted); font-weight: 600; padding: 12px 18px;
    background: var(--surface-2); border-bottom: 1px solid var(--border); white-space: nowrap;
  }
  tbody td { padding: 14px 18px; border-bottom: 1px solid var(--border); vertical-align: middle; }
  tbody tr:last-child td { border-bottom: none; }
  tbody tr { transition: background .12s ease; }
  tbody tr:hover { background: var(--surface-2); }
  tbody tr.current { background: var(--current); box-shadow: inset 3px 0 0 var(--current-bar); }
  tbody tr.current:hover { background: var(--current); }
  .id-cell a { font-family: var(--mono); font-size: 12.5px; color: var(--text); }
  .id-cell a:hover { color: var(--brand); text-decoration: underline; }
  .node-cell { display: flex; align-items: center; gap: 8px; font-weight: 550; }
  .addr { font-family: var(--mono); font-size: 12.5px; color: var(--text-muted); }

  /* Chips & badges */
  .chip {
    display: inline-flex; align-items: center; gap: 5px; font-size: 11px; font-weight: 600;
    padding: 3px 9px; border-radius: 999px; line-height: 1.4; white-space: nowrap;
  }
  .chip-you { background: var(--brand-soft); color: var(--brand); border: 1px solid transparent; }
  .badge {
    display: inline-flex; align-items: center; gap: 6px; font-size: 12px; font-weight: 600;
    padding: 4px 11px; border-radius: 999px; white-space: nowrap; border: 1px solid transparent;
  }
  .badge .dot { width: 7px; height: 7px; border-radius: 50%; flex: none; }
  .badge.passing { background: var(--ok-soft); color: var(--ok); border-color: var(--ok-border); }
  .badge.passing .dot { background: var(--ok); }
  .badge.warning { background: var(--warn-soft); color: var(--warn); border-color: var(--warn-border); }
  .badge.warning .dot { background: var(--warn); }
  .badge.critical { background: var(--crit-soft); color: var(--crit); border-color: var(--crit-border); }
  .badge.critical .dot { background: var(--crit); }
  .badge.maintenance,
  .badge.unknown { background: var(--muted-soft); color: var(--text-muted); border-color: var(--muted-border); }
  .badge.maintenance .dot,
  .badge.unknown .dot { background: var(--text-faint); }

  .leader-badge {
    display: inline-flex; align-items: center; gap: 6px; font-size: 12px; font-weight: 650;
    padding: 4px 11px; border-radius: 999px;
    background: var(--leader-soft); color: var(--leader); border: 1px solid var(--leader-border);
  }
  .dash { color: var(--text-faint); }

  /* Detail page */
  .detail-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 0; }
  @media (max-width: 720px) { .detail-grid { grid-template-columns: 1fr; } }
  .field { padding: 16px 20px; border-bottom: 1px solid var(--border); border-right: 1px solid var(--border); }
  .detail-grid .field:nth-child(2n) { border-right: none; }
  .field .k { font-size: 11px; text-transform: uppercase; letter-spacing: .05em; color: var(--text-muted); font-weight: 600; }
  .field .v { margin-top: 6px; font-size: 14px; font-weight: 520; word-break: break-word; }
  .field .v.mono { font-family: var(--mono); font-size: 13px; }
  .full { grid-column: 1 / -1; border-right: none; }

  .tags { display: flex; flex-wrap: wrap; gap: 7px; margin-top: 8px; }
  .tag { font-family: var(--mono); font-size: 11.5px; background: var(--muted-soft); color: var(--text-muted);
    border: 1px solid var(--muted-border); border-radius: 7px; padding: 3px 8px; }

  .kv-tbl { width: 100%; border-collapse: collapse; font-size: 13px; }
  .kv-tbl td { padding: 10px 20px; border-bottom: 1px solid var(--border); }
  .kv-tbl tr:last-child td { border-bottom: none; }
  .kv-tbl td.k { color: var(--text-muted); font-weight: 600; width: 220px; font-size: 12px;
    text-transform: uppercase; letter-spacing: .04em; }
  .kv-tbl td.v { font-family: var(--mono); font-size: 12.5px; word-break: break-word; }

  .btn {
    display: inline-flex; align-items: center; gap: 7px; font-size: 13px; font-weight: 550;
    padding: 8px 14px; border-radius: 9px; border: 1px solid var(--border);
    background: var(--surface); color: var(--text); cursor: pointer;
  }
  .btn:hover { background: var(--surface-2); }
  .btn-primary { background: var(--brand); color: #fff; border-color: transparent; }
  .btn-primary:hover { filter: brightness(1.05); background: var(--brand); }

  .empty { padding: 48px 20px; text-align: center; color: var(--text-muted); }
  .empty .big { font-size: 15px; font-weight: 600; color: var(--text); margin-bottom: 4px; }

  .footer { margin-top: 30px; font-size: 12px; color: var(--text-faint); display: flex;
    flex-wrap: wrap; gap: 6px 16px; align-items: center; }
  .footer .dotsep { color: var(--border); }

  /* Error page */
  .error-wrap { min-height: 78vh; display: grid; place-items: center; padding: 40px 24px; }
  .error-card { max-width: 520px; width: 100%; text-align: center; background: var(--surface);
    border: 1px solid var(--border); border-radius: 18px; box-shadow: var(--shadow); padding: 44px 40px; }
  .error-code { font-size: 68px; font-weight: 750; letter-spacing: -.03em; line-height: 1;
    background: linear-gradient(135deg, var(--brand), #a855f7); -webkit-background-clip: text;
    background-clip: text; color: transparent; }
  .error-tag { display: inline-block; margin-top: 14px; font-family: var(--mono); font-size: 12px;
    color: var(--crit); background: var(--crit-soft); border: 1px solid var(--crit-border);
    padding: 3px 10px; border-radius: 999px; }
  .error-msg { font-size: 19px; font-weight: 640; margin-top: 18px; letter-spacing: -.01em; }
  .error-hint { color: var(--text-muted); font-size: 14px; margin-top: 8px; }
  .error-path { font-family: var(--mono); font-size: 12px; color: var(--text-faint);
    margin-top: 20px; background: var(--surface-2); border: 1px solid var(--border);
    border-radius: 8px; padding: 8px 12px; word-break: break-all; }
  .error-actions { margin-top: 26px; display: flex; gap: 10px; justify-content: center; }
</style>
{{end}}

{{define "footer"}}
<div class="footer">
  <span>{{.AppName}} <b>{{.Version}}</b></span>
  <span class="dotsep">&bull;</span>
  <span>rev {{.Revision}}</span>
  <span class="dotsep">&bull;</span>
  <span>generated {{.GeneratedAt}}</span>
</div>
{{end}}


{{define "peers"}}<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta http-equiv="refresh" content="15">
  <title>{{.AppName}} &middot; Cluster Peers</title>
  {{template "styles" .}}
</head>
<body>
  <header class="masthead">
    <div class="wrap">
      <div class="brandline">
        <div class="logo">{{if .AppName}}{{slice .AppName 0 1}}{{else}}C{{end}}</div>
        <div>
          <h1>{{.AppName}} Cluster</h1>
          <div class="sub">Service discovery &amp; leader election dashboard</div>
        </div>
      </div>
      <div class="crumbs">Node <span style="opacity:.5">/</span> Peers</div>
      <div class="stats">
        <div class="stat">
          <div class="label">Nodes</div>
          <div class="value">{{.Total}}</div>
        </div>
        <div class="stat">
          <div class="label">Healthy</div>
          <div class="value">{{.Healthy}}<span style="font-size:13px;color:#94a3b8;font-weight:500">/ {{.Total}}</span></div>
        </div>
        <div class="stat">
          <div class="label">Leader</div>
          {{if not .ElectionOn}}
            <div class="value small" style="color:#94a3b8">Election off</div>
          {{else if .LeaderKnown}}
            <div class="value small">&#9819; {{.LeaderID}}</div>
          {{else}}
            <div class="value small" style="color:#fbbf24">No leader</div>
          {{end}}
        </div>
      </div>
    </div>
  </header>

  <div class="wrap">
    <div class="card">
      <div class="card-head">
        <h2>Cluster Peers</h2>
        <span class="meta">Auto-refreshes every 15s</span>
      </div>
      <div class="tbl-scroll">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>Node</th>
              <th>HTTP Address</th>
              <th>Status</th>
              <th>Leadership</th>
            </tr>
          </thead>
          <tbody>
            {{range .Rows}}
            <tr class="{{if .IsCurrent}}current{{end}}">
              <td class="id-cell">
                <div class="node-cell">
                  <a href="{{$.PeerBasePath}}/{{.ID}}">{{.ID}}</a>
                  {{if .IsCurrent}}<span class="chip chip-you">&#9679; This node</span>{{end}}
                </div>
              </td>
              <td>{{.Node}}</td>
              <td><span class="addr">{{httpAddr .Peer}}</span></td>
              <td>
                <span class="badge {{statusClass .Status}}"><span class="dot"></span>{{statusLabel .Status}}</span>
              </td>
              <td>
                {{if .IsLeader}}<span class="leader-badge">&#9819; Leader</span>{{else}}<span class="dash">&mdash;</span>{{end}}
              </td>
            </tr>
            {{else}}
            <tr><td colspan="5">
              <div class="empty"><div class="big">No peers discovered</div>Waiting for other nodes to register with Consul.</div>
            </td></tr>
            {{end}}
          </tbody>
        </table>
      </div>
    </div>
    {{template "footer" .}}
  </div>
</body>
</html>{{end}}


{{define "peer"}}<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.AppName}} &middot; {{.Row.Node}}</title>
  {{template "styles" .}}
</head>
<body>
  <header class="masthead">
    <div class="wrap">
      <div class="brandline">
        <div class="logo">{{if .AppName}}{{slice .AppName 0 1}}{{else}}C{{end}}</div>
        <div>
          <h1>{{.Row.Node}}</h1>
          <div class="sub">Peer detail</div>
        </div>
      </div>
      <div class="crumbs"><a href="{{.ListPath}}">Peers</a> <span style="opacity:.5">/</span> {{.Row.ID}}</div>
      <div class="stats">
        <div class="stat">
          <div class="label">Status</div>
          <div class="value small">{{statusLabel .Row.Status}}</div>
        </div>
        <div class="stat">
          <div class="label">Role</div>
          <div class="value small">{{if .Row.IsCurrent}}This node{{else}}Peer{{end}}</div>
        </div>
        <div class="stat">
          <div class="label">Leadership</div>
          {{if not .ElectionOn}}
            <div class="value small" style="color:#94a3b8">Election off</div>
          {{else if .Row.IsLeader}}
            <div class="value small">&#9819; Leader</div>
          {{else}}
            <div class="value small" style="color:#94a3b8">Follower</div>
          {{end}}
        </div>
      </div>
    </div>
  </header>

  <div class="wrap">
    <div class="card">
      <div class="card-head">
        <h2>Overview</h2>
        <span class="meta">
          <span class="badge {{statusClass .Row.Status}}"><span class="dot"></span>{{statusLabel .Row.Status}}</span>
          {{if and .ElectionOn .Row.IsLeader}}<span class="leader-badge" style="margin-left:6px">&#9819; Leader</span>{{end}}
        </span>
      </div>
      <div class="detail-grid">
        <div class="field full">
          <div class="k">Service ID</div>
          <div class="v mono">{{.Row.ID}}{{if .Row.IsCurrent}} <span class="chip chip-you" style="margin-left:6px">&#9679; This node</span>{{end}}</div>
        </div>
        <div class="field">
          <div class="k">Node</div>
          <div class="v">{{.Row.Node}}</div>
        </div>
        <div class="field">
          <div class="k">Status</div>
          <div class="v"><span class="badge {{statusClass .Row.Status}}"><span class="dot"></span>{{statusLabel .Row.Status}}</span></div>
        </div>
        <div class="field">
          <div class="k">Address</div>
          <div class="v mono">{{.Row.Address}}</div>
        </div>
        <div class="field">
          <div class="k">HTTP Address</div>
          <div class="v mono">{{httpAddr .Row.Peer}}</div>
        </div>
        <div class="field">
          <div class="k">HTTP Port</div>
          <div class="v mono">{{.Row.HTTPPort}}</div>
        </div>
        <div class="field">
          <div class="k">Is Leader</div>
          <div class="v">{{if not .ElectionOn}}<span class="dash">Election disabled</span>{{else if .Row.IsLeader}}<span class="leader-badge">&#9819; Yes</span>{{else}}<span class="dash">No</span>{{end}}</div>
        </div>
        <div class="field full">
          <div class="k">Tags</div>
          {{if .Row.Tags}}<div class="tags">{{range .Row.Tags}}<span class="tag">{{.}}</span>{{end}}</div>{{else}}<div class="v"><span class="dash">&mdash;</span></div>{{end}}
        </div>
      </div>
    </div>

    <div class="card">
      <div class="card-head"><h2>Metadata</h2><span class="meta">{{len .Meta}} entries</span></div>
      {{if .Meta}}
      <table class="kv-tbl">
        <tbody>
          {{range .Meta}}
          <tr><td class="k">{{.Key}}</td><td class="v">{{.Value}}</td></tr>
          {{end}}
        </tbody>
      </table>
      {{else}}
      <div class="empty">No metadata reported for this peer.</div>
      {{end}}
    </div>

    <div style="margin-top:22px">
      <a class="btn" href="{{.ListPath}}">&larr; Back to peers</a>
    </div>
    {{template "footer" .}}
  </div>
</body>
</html>{{end}}


{{define "error"}}<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.AppName}} &middot; Error {{.Status}}</title>
  {{template "styles" .}}
</head>
<body>
  <div class="error-wrap">
    <div class="error-card">
      <div class="error-code">{{.Status}}</div>
      <div class="error-tag">{{.Code}}</div>
      <div class="error-msg">{{.Message}}</div>
      <div class="error-hint">{{.Hint}}</div>
      <div class="error-path">{{.Path}}</div>
      <div class="error-actions">
        <a class="btn btn-primary" href="/">&larr; Back to dashboard</a>
      </div>
    </div>
  </div>
</body>
</html>{{end}}
`
