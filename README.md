import { useState, useEffect, useRef, useCallback } from "react";
import { LineChart, Line, BarChart, Bar, AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ReferenceLine } from "recharts";

// ─── COLOR SYSTEM ───────────────────────────────────────────────────────────
const C = {
  bg: "#08090c",
  bg1: "#0d0f14",
  bg2: "#121520",
  bg3: "#171b26",
  border: "rgba(255,255,255,0.06)",
  borderHi: "rgba(255,255,255,0.12)",
  cyan: "#00f5c4",
  cyanDim: "rgba(0,245,196,0.15)",
  cyanGlow: "rgba(0,245,196,0.08)",
  blue: "#4f9eff",
  blueDim: "rgba(79,158,255,0.12)",
  amber: "#ffb84d",
  amberDim: "rgba(255,184,77,0.12)",
  red: "#ff5757",
  redDim: "rgba(255,87,87,0.12)",
  purple: "#b48aff",
  purpleDim: "rgba(180,138,255,0.12)",
  green: "#4ade80",
  greenDim: "rgba(74,222,128,0.12)",
  text: "#e8eaf0",
  textMuted: "#6b7280",
  textDim: "#3d4252",
};

// ─── GLOBAL STYLES ───────────────────────────────────────────────────────────
const globalStyle = `
@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@300;400;500;700&family=Syne:wght@400;600;700;800&family=DM+Sans:wght@300;400;500&display=swap');

* { box-sizing: border-box; margin: 0; padding: 0; }
body { background: ${C.bg}; color: ${C.text}; font-family: 'DM Sans', sans-serif; }

@keyframes pulse-ring {
  0% { transform: scale(0.8); opacity: 0.8; }
  100% { transform: scale(2.2); opacity: 0; }
}
@keyframes flow-dash {
  0% { stroke-dashoffset: 0; }
  100% { stroke-dashoffset: -80; }
}
@keyframes fade-up {
  from { opacity: 0; transform: translateY(16px); }
  to { opacity: 1; transform: translateY(0); }
}
@keyframes glow-pulse {
  0%, 100% { opacity: 0.5; }
  50% { opacity: 1; }
}
@keyframes scan-line {
  0% { transform: translateY(0%); }
  100% { transform: translateY(100%); }
}
@keyframes counter-tick {
  0% { transform: translateY(0); }
  50% { transform: translateY(-50%); }
  100% { transform: translateY(-100%); }
}
@keyframes bar-fill {
  from { transform: scaleX(0); }
  to { transform: scaleX(1); }
}
@keyframes ring-spin {
  from { stroke-dashoffset: 220; }
  to { stroke-dashoffset: 0; }
}
@keyframes packet {
  0% { opacity: 0; transform: translateX(-8px); }
  20% { opacity: 1; }
  80% { opacity: 1; }
  100% { opacity: 0; transform: translateX(calc(100% + 8px)); }
}
@keyframes worker-pulse {
  0%, 100% { background: ${C.bg2}; box-shadow: none; }
  50% { background: ${C.cyanDim}; box-shadow: 0 0 12px ${C.cyanDim}; }
}
@keyframes queue-fill {
  0% { height: 0%; opacity: 0; }
  100% { height: var(--fill); opacity: 1; }
}
@keyframes latency-spike {
  0%, 100% { transform: scaleY(1); }
  50% { transform: scaleY(1.3); }
}
@keyframes shimmer {
  0% { background-position: -200% 0; }
  100% { background-position: 200% 0; }
}

.section-reveal { animation: fade-up 0.6s ease forwards; }
.mono { font-family: 'JetBrains Mono', monospace; }
.syne { font-family: 'Syne', sans-serif; }

.neon-text-cyan { color: ${C.cyan}; text-shadow: 0 0 20px rgba(0,245,196,0.4); }
.neon-text-blue { color: ${C.blue}; text-shadow: 0 0 20px rgba(79,158,255,0.4); }
.neon-border { border: 1px solid ${C.cyanDim}; box-shadow: 0 0 20px ${C.cyanGlow} inset, 0 0 20px ${C.cyanGlow}; }

.glass {
  background: rgba(13,15,20,0.8);
  backdrop-filter: blur(12px);
  border: 1px solid ${C.border};
  border-radius: 12px;
}

::-webkit-scrollbar { width: 4px; }
::-webkit-scrollbar-track { background: ${C.bg}; }
::-webkit-scrollbar-thumb { background: ${C.bg3}; border-radius: 2px; }

.tooltip-custom {
  background: ${C.bg2} !important;
  border: 1px solid ${C.border} !important;
  border-radius: 8px !important;
  font-family: 'JetBrains Mono', monospace !important;
  font-size: 11px !important;
  color: ${C.text} !important;
  padding: 8px 12px !important;
}
`;

// ─── UTILS ───────────────────────────────────────────────────────────────────
const lerp = (a, b, t) => a + (b - a) * t;
const clamp = (v, lo, hi) => Math.max(lo, Math.min(hi, v));
const fmt = (n) => n >= 1e9 ? `${(n / 1e9).toFixed(1)}B` : n >= 1e6 ? `${(n / 1e6).toFixed(1)}M` : n >= 1e3 ? `${(n / 1e3).toFixed(1)}K` : String(n);

// ─── ANIMATED COUNTER ────────────────────────────────────────────────────────
function AnimCounter({ value, suffix = "", duration = 1800 }) {
  const [display, setDisplay] = useState(0);
  const ref = useRef(null);
  useEffect(() => {
    let start = null;
    const from = 0;
    const step = (ts) => {
      if (!start) start = ts;
      const p = Math.min((ts - start) / duration, 1);
      const ease = 1 - Math.pow(1 - p, 3);
      setDisplay(Math.round(lerp(from, value, ease)));
      if (p < 1) ref.current = requestAnimationFrame(step);
    };
    ref.current = requestAnimationFrame(step);
    return () => cancelAnimationFrame(ref.current);
  }, [value, duration]);
  return <span>{fmt(display)}{suffix}</span>;
}

// ─── SECTION WRAPPER ─────────────────────────────────────────────────────────
function Section({ id, children, style = {} }) {
  return (
    <section id={id} style={{ padding: "80px 0", ...style }}>
      {children}
    </section>
  );
}

// ─── PILL BADGE ──────────────────────────────────────────────────────────────
function Badge({ children, color = C.cyan }) {
  return (
    <span className="mono" style={{
      fontSize: 10, fontWeight: 500, letterSpacing: "0.1em",
      padding: "3px 10px", borderRadius: 99,
      border: `1px solid ${color}30`,
      color, background: `${color}10`,
      textTransform: "uppercase",
    }}>{children}</span>
  );
}

// ─── METRIC CARD ─────────────────────────────────────────────────────────────
function MetricCard({ label, value, unit, color = C.cyan, sub, live = false }) {
  const [v, setV] = useState(parseFloat(value));
  useEffect(() => {
    if (!live) return;
    const id = setInterval(() => setV(prev => +(prev + (Math.random() - 0.5) * 0.3).toFixed(2)), 800);
    return () => clearInterval(id);
  }, [live]);
  return (
    <div className="glass" style={{ padding: "20px 24px", position: "relative", overflow: "hidden" }}>
      <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: 1, background: `linear-gradient(90deg, transparent, ${color}80, transparent)` }} />
      <div style={{ fontSize: 11, color: C.textMuted, fontFamily: "'JetBrains Mono'", letterSpacing: "0.08em", textTransform: "uppercase", marginBottom: 8 }}>{label}</div>
      <div style={{ fontSize: 28, fontFamily: "'Syne'", fontWeight: 700, color }}>
        {live ? v.toFixed(1) : value}
        <span style={{ fontSize: 14, fontWeight: 400, color: C.textMuted, marginLeft: 4 }}>{unit}</span>
      </div>
      {sub && <div style={{ fontSize: 11, color: C.textMuted, marginTop: 4, fontFamily: "'JetBrains Mono'" }}>{sub}</div>}
      {live && <div style={{ position: "absolute", top: 12, right: 12, width: 6, height: 6, borderRadius: "50%", background: color, animation: "glow-pulse 1.2s ease-in-out infinite" }} />}
    </div>
  );
}

// ─── ARCHITECTURE DIAGRAM ────────────────────────────────────────────────────
function ArchDiagram() {
  const [active, setActive] = useState(null);
  const [animPhase, setAnimPhase] = useState(0);

  useEffect(() => {
    const id = setInterval(() => setAnimPhase(p => (p + 1) % 6), 900);
    return () => clearInterval(id);
  }, []);

  const nodes = [
    { id: 0, label: "Client", sub: "HTTP / gRPC", x: 50, y: 180, color: C.purple },
    { id: 1, label: "Dispatcher", sub: "FNV-1a routing", x: 200, y: 180, color: C.blue },
    { id: 2, label: "RingBuffer", sub: "lock-free queue", x: 360, y: 120, color: C.cyan },
    { id: 3, label: "RingBuffer", sub: "lock-free queue", x: 360, y: 180, color: C.cyan },
    { id: 4, label: "RingBuffer", sub: "lock-free queue", x: 360, y: 240, color: C.cyan },
    { id: 5, label: "Workers", sub: "parallel execution", x: 520, y: 180, color: C.amber },
    { id: 6, label: "Aggregator", sub: "microbatch lanes", x: 660, y: 180, color: C.green },
    { id: 7, label: "Redis Lua", sub: "atomic pipeline", x: 800, y: 180, color: C.red },
  ];

  const edges = [
    [0, 1], [1, 2], [1, 3], [1, 4],
    [2, 5], [3, 5], [4, 5],
    [5, 6], [6, 7],
  ];

  const W = 920, H = 340;
  const nodeW = 100, nodeH = 48;

  return (
    <div style={{ width: "100%", overflowX: "auto", padding: "24px 0" }}>
      <svg viewBox={`0 0 ${W} ${H}`} style={{ width: "100%", minWidth: 600, display: "block" }}>
        <defs>
          <marker id="ah" markerWidth="6" markerHeight="6" refX="5" refY="3" orient="auto">
            <path d="M0,0 L6,3 L0,6" fill="none" stroke="rgba(255,255,255,0.25)" strokeWidth="1" />
          </marker>
          {[C.cyan, C.blue, C.amber, C.green, C.red, C.purple].map((c, i) => (
            <filter key={i} id={`glow-${i}`}>
              <feGaussianBlur stdDeviation="3" result="blur" />
              <feMerge><feMergeNode in="blur" /><feMergeNode in="SourceGraphic" /></feMerge>
            </filter>
          ))}
        </defs>

        {/* Grid lines */}
        {[80, 160, 240].map(y => (
          <line key={y} x1={0} y1={y} x2={W} y2={y} stroke="rgba(255,255,255,0.02)" strokeWidth="1" />
        ))}

        {/* Edges */}
        {edges.map(([a, b], ei) => {
          const na = nodes[a], nb = nodes[b];
          const x1 = na.x + nodeW, y1 = na.y + nodeH / 2;
          const x2 = nb.x, y2 = nb.y + nodeH / 2;
          const isActive = animPhase === ei % 6;
          return (
            <g key={ei}>
              <path
                d={`M${x1},${y1} C${x1 + 30},${y1} ${x2 - 30},${y2} ${x2},${y2}`}
                fill="none"
                stroke={isActive ? "rgba(255,255,255,0.3)" : "rgba(255,255,255,0.08)"}
                strokeWidth={isActive ? 1.5 : 1}
                strokeDasharray="6 4"
                style={{ animation: "flow-dash 1.5s linear infinite" }}
                markerEnd="url(#ah)"
              />
              {isActive && (
                <circle r={3} fill={C.cyan} opacity={0.9}>
                  <animateMotion dur="0.9s" repeatCount="indefinite"
                    path={`M${x1},${y1} C${x1 + 30},${y1} ${x2 - 30},${y2} ${x2},${y2}`} />
                </circle>
              )}
            </g>
          );
        })}

        {/* Nodes */}
        {nodes.map(n => (
          <g key={n.id} style={{ cursor: "pointer" }} onClick={() => setActive(active === n.id ? null : n.id)}>
            <rect
              x={n.x} y={n.y} width={nodeW} height={nodeH} rx={8}
              fill={active === n.id ? `${n.color}20` : C.bg2}
              stroke={active === n.id ? n.color : `${n.color}50`}
              strokeWidth={active === n.id ? 1.5 : 1}
            />
            <text x={n.x + nodeW / 2} y={n.y + 17} textAnchor="middle"
              fill={n.color} fontSize={11} fontFamily="'JetBrains Mono'" fontWeight={500}>{n.label}</text>
            <text x={n.x + nodeW / 2} y={n.y + 32} textAnchor="middle"
              fill={C.textMuted} fontSize={9} fontFamily="'JetBrains Mono'">{n.sub}</text>
          </g>
        ))}
      </svg>
    </div>
  );
}

// ─── LOCK-FREE RINGBUFFER VIZ ─────────────────────────────────────────────────
function RingBufferViz() {
  const N = 16;
  const [head, setHead] = useState(0);
  const [tail, setTail] = useState(5);
  const [cells, setCells] = useState(() => Array.from({ length: N }, (_, i) => ({ filled: i < 5, id: i })));

  useEffect(() => {
    const id = setInterval(() => {
      setHead(h => {
        const nh = (h + 1) % N;
        setTail(t => {
          const nt = (t + 1) % N;
          setCells(prev => {
            const next = [...prev];
            next[nt] = { ...next[nt], filled: true, flash: true };
            next[h] = { ...next[h], filled: false, flash: false };
            setTimeout(() => setCells(c => {
              const a = [...c]; a[nt] = { ...a[nt], flash: false }; return a;
            }), 200);
            return next;
          });
          return nt;
        });
        return nh;
      });
    }, 400);
    return () => clearInterval(id);
  }, []);

  const r = 90, cx = 130, cy = 130;
  return (
    <div style={{ display: "flex", gap: 24, flexWrap: "wrap", alignItems: "center" }}>
      <svg viewBox="0 0 260 260" style={{ width: 220, flexShrink: 0 }}>
        {cells.map((cell, i) => {
          const angle = (i / N) * Math.PI * 2 - Math.PI / 2;
          const x = cx + r * Math.cos(angle);
          const y = cy + r * Math.sin(angle);
          const isHead = i === head, isTail = i === tail;
          return (
            <g key={i}>
              <rect x={x - 10} y={y - 10} width={20} height={20} rx={3}
                fill={cell.filled ? (cell.flash ? C.cyan : `${C.cyan}60`) : C.bg3}
                stroke={isHead ? C.amber : isTail ? C.purple : cell.filled ? `${C.cyan}40` : C.textDim}
                strokeWidth={isHead || isTail ? 2 : 0.5}
                style={{ transition: "fill 0.15s" }}
              />
              {(isHead || isTail) && (
                <text x={x} y={y + 4} textAnchor="middle" fill={isHead ? C.amber : C.purple}
                  fontSize={7} fontFamily="'JetBrains Mono'" fontWeight={700}>
                  {isHead ? "H" : "T"}
                </text>
              )}
            </g>
          );
        })}
        <text x={cx} y={cy - 10} textAnchor="middle" fill={C.textMuted} fontSize={9} fontFamily="'JetBrains Mono'">RING</text>
        <text x={cx} y={cy + 8} textAnchor="middle" fill={C.cyan} fontSize={14} fontFamily="'Syne'" fontWeight={700}>
          {cells.filter(c => c.filled).length}
        </text>
        <text x={cx} y={cy + 22} textAnchor="middle" fill={C.textMuted} fontSize={8} fontFamily="'JetBrains Mono'">
          /{N} slots
        </text>
      </svg>
      <div style={{ flex: 1, minWidth: 160 }}>
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          {[
            { label: "HEAD (consumer)", color: C.amber, desc: "non-blocking dequeue" },
            { label: "TAIL (producer)", color: C.purple, desc: "atomic CAS enqueue" },
            { label: "FILLED slot", color: C.cyan, desc: "zero-alloc hotpath" },
            { label: "EMPTY slot", color: C.textDim, desc: "pre-allocated buffer" },
          ].map(({ label, color, desc }) => (
            <div key={label} style={{ display: "flex", alignItems: "center", gap: 8 }}>
              <div style={{ width: 10, height: 10, borderRadius: 2, background: color, flexShrink: 0 }} />
              <div>
                <div className="mono" style={{ fontSize: 10, color, fontWeight: 500 }}>{label}</div>
                <div className="mono" style={{ fontSize: 9, color: C.textMuted }}>{desc}</div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

// ─── THROUGHPUT SCALING CHART ────────────────────────────────────────────────
const scalingData = [
  { workers: 2, throughput: 5.5, latency: 211.3 },
  { workers: 4, throughput: 6.4, latency: 178.0 },
  { workers: 8, throughput: 10.3, latency: 116.3 },
  { workers: 16, throughput: 11.4, latency: 106.2 },
];

function ScalingChart() {
  return (
    <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
      <div className="glass" style={{ padding: 20 }}>
        <div className="mono" style={{ fontSize: 10, color: C.textMuted, marginBottom: 12, textTransform: "uppercase", letterSpacing: "0.08em" }}>throughput / workers</div>
        <ResponsiveContainer width="100%" height={180}>
          <AreaChart data={scalingData} margin={{ top: 5, right: 10, bottom: 5, left: 0 }}>
            <defs>
              <linearGradient id="tGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={C.cyan} stopOpacity={0.3} />
                <stop offset="100%" stopColor={C.cyan} stopOpacity={0.01} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.04)" />
            <XAxis dataKey="workers" tick={{ fill: C.textMuted, fontSize: 10, fontFamily: "'JetBrains Mono'" }} />
            <YAxis tick={{ fill: C.textMuted, fontSize: 10, fontFamily: "'JetBrains Mono'" }} unit="M" />
            <Tooltip contentStyle={{ background: C.bg2, border: `1px solid ${C.border}`, borderRadius: 8, fontFamily: "'JetBrains Mono'", fontSize: 11 }} />
            <Area type="monotone" dataKey="throughput" stroke={C.cyan} strokeWidth={2} fill="url(#tGrad)" dot={{ fill: C.cyan, r: 4 }} />
          </AreaChart>
        </ResponsiveContainer>
      </div>
      <div className="glass" style={{ padding: 20 }}>
        <div className="mono" style={{ fontSize: 10, color: C.textMuted, marginBottom: 12, textTransform: "uppercase", letterSpacing: "0.08em" }}>latency (ns/op) / workers</div>
        <ResponsiveContainer width="100%" height={180}>
          <AreaChart data={scalingData} margin={{ top: 5, right: 10, bottom: 5, left: 0 }}>
            <defs>
              <linearGradient id="lGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={C.amber} stopOpacity={0.3} />
                <stop offset="100%" stopColor={C.amber} stopOpacity={0.01} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.04)" />
            <XAxis dataKey="workers" tick={{ fill: C.textMuted, fontSize: 10, fontFamily: "'JetBrains Mono'" }} />
            <YAxis tick={{ fill: C.textMuted, fontSize: 10, fontFamily: "'JetBrains Mono'" }} unit="ns" />
            <Tooltip contentStyle={{ background: C.bg2, border: `1px solid ${C.border}`, borderRadius: 8, fontFamily: "'JetBrains Mono'", fontSize: 11 }} />
            <Area type="monotone" dataKey="latency" stroke={C.amber} strokeWidth={2} fill="url(#lGrad)" dot={{ fill: C.amber, r: 4 }} />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

// ─── QUEUE SATURATION VIZ ───────────────────────────────────────────────────
function QueueSaturation() {
  const [time, setTime] = useState(0);
  const [running, setRunning] = useState(true);

  useEffect(() => {
    if (!running) return;
    const id = setInterval(() => setTime(t => (t + 1) % 120), 80);
    return () => clearInterval(id);
  }, [running]);

  const shards = 8;
  const queues = Array.from({ length: shards }, (_, i) => {
    const phase = (time + i * 11) % 120;
    const fill = Math.min(1, Math.max(0.05, 0.3 + 0.5 * Math.sin(phase / 120 * Math.PI * 2) + (i % 3 === 0 ? 0.3 * Math.sin(phase / 30 * Math.PI) : 0)));
    const isHot = i === 2 || i === 5;
    return { fill: isHot ? Math.min(1, fill * 1.4) : fill, isHot };
  });

  const historyLen = 60;
  const history = Array.from({ length: historyLen }, (_, j) => {
    const t2 = (time - historyLen + j + 120) % 120;
    return {
      t: j,
      depth: Math.round(clamp(40 + 60 * Math.sin(t2 / 120 * Math.PI * 2) + 20 * Math.sin(t2 / 20 * Math.PI), 5, 100)),
      rejected: Math.round(clamp((t2 > 60 ? (t2 - 60) * 1.5 : 0), 0, 80)),
    };
  });

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      {/* Shard queue bars */}
      <div className="glass" style={{ padding: 20 }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 16 }}>
          <div className="mono" style={{ fontSize: 10, color: C.textMuted, textTransform: "uppercase", letterSpacing: "0.08em" }}>shard queue depth (live)</div>
          <button onClick={() => setRunning(r => !r)}
            style={{ background: "none", border: `1px solid ${C.border}`, borderRadius: 6, padding: "4px 10px", color: C.textMuted, fontSize: 10, cursor: "pointer", fontFamily: "'JetBrains Mono'" }}>
            {running ? "pause" : "resume"}
          </button>
        </div>
        <div style={{ display: "flex", gap: 8, alignItems: "flex-end", height: 80 }}>
          {queues.map((q, i) => (
            <div key={i} style={{ flex: 1, display: "flex", flexDirection: "column", alignItems: "center", gap: 4 }}>
              <div style={{ width: "100%", background: C.bg3, borderRadius: 3, height: 64, position: "relative", overflow: "hidden" }}>
                <div style={{
                  position: "absolute", bottom: 0, left: 0, right: 0,
                  height: `${q.fill * 100}%`,
                  background: q.fill > 0.8 ? C.red : q.fill > 0.6 ? C.amber : C.cyan,
                  transition: "height 0.12s ease, background 0.3s",
                  borderRadius: "0 0 3px 3px",
                }} />
                {q.isHot && (
                  <div className="mono" style={{ position: "absolute", top: 2, left: 0, right: 0, textAlign: "center", fontSize: 7, color: C.red, fontWeight: 700 }}>HOT</div>
                )}
              </div>
              <div className="mono" style={{ fontSize: 8, color: C.textMuted }}>S{i}</div>
            </div>
          ))}
        </div>
      </div>

      {/* History */}
      <div className="glass" style={{ padding: 20 }}>
        <div className="mono" style={{ fontSize: 10, color: C.textMuted, marginBottom: 12, textTransform: "uppercase", letterSpacing: "0.08em" }}>queue depth + rejection rate (60s window)</div>
        <ResponsiveContainer width="100%" height={140}>
          <AreaChart data={history} margin={{ top: 5, right: 10, bottom: 0, left: 0 }}>
            <defs>
              <linearGradient id="qGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={C.blue} stopOpacity={0.4} />
                <stop offset="100%" stopColor={C.blue} stopOpacity={0.01} />
              </linearGradient>
              <linearGradient id="rGrad" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={C.red} stopOpacity={0.4} />
                <stop offset="100%" stopColor={C.red} stopOpacity={0.01} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.03)" />
            <XAxis dataKey="t" hide />
            <YAxis tick={{ fill: C.textMuted, fontSize: 9, fontFamily: "'JetBrains Mono'" }} />
            <Tooltip contentStyle={{ background: C.bg2, border: `1px solid ${C.border}`, borderRadius: 8, fontFamily: "'JetBrains Mono'", fontSize: 10 }} />
            <Area type="monotone" dataKey="depth" name="Queue depth" stroke={C.blue} strokeWidth={1.5} fill="url(#qGrad)" />
            <Area type="monotone" dataKey="rejected" name="Rejected %" stroke={C.red} strokeWidth={1.5} fill="url(#rGrad)" />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}

// ─── LATENCY PERCENTILE CHART ─────────────────────────────────────────────────
function LatencyChart() {
  const [tick, setTick] = useState(0);
  useEffect(() => {
    const id = setInterval(() => setTick(t => t + 1), 600);
    return () => clearInterval(id);
  }, []);

  const baseP50 = 115, baseP95 = 142, baseP99 = 165;
  const data = Array.from({ length: 40 }, (_, i) => {
    const jitter = Math.sin(i * 0.4 + tick * 0.15) * 8;
    const spike = i === 22 ? 30 : i === 35 ? 20 : 0;
    return {
      t: i,
      p50: Math.round(baseP50 + jitter * 0.3 + spike * 0.2),
      p95: Math.round(baseP95 + jitter * 0.7 + spike * 0.6),
      p99: Math.round(baseP99 + jitter + spike),
    };
  });

  return (
    <div className="glass" style={{ padding: 24 }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 16 }}>
        <div className="mono" style={{ fontSize: 10, color: C.textMuted, textTransform: "uppercase", letterSpacing: "0.08em" }}>latency percentiles — ns/op (simulated live)</div>
        <div style={{ display: "flex", gap: 12 }}>
          {[{ l: "p50", c: C.cyan }, { l: "p95", c: C.amber }, { l: "p99", c: C.red }].map(({ l, c }) => (
            <div key={l} style={{ display: "flex", alignItems: "center", gap: 4 }}>
              <div style={{ width: 16, height: 2, background: c, borderRadius: 1 }} />
              <span className="mono" style={{ fontSize: 9, color: c }}>{l}</span>
            </div>
          ))}
        </div>
      </div>
      <ResponsiveContainer width="100%" height={200}>
        <LineChart data={data} margin={{ top: 5, right: 10, bottom: 5, left: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.04)" />
          <XAxis dataKey="t" hide />
          <YAxis tick={{ fill: C.textMuted, fontSize: 10, fontFamily: "'JetBrains Mono'" }} unit="ns" domain={[80, 220]} />
          <Tooltip contentStyle={{ background: C.bg2, border: `1px solid ${C.border}`, borderRadius: 8, fontFamily: "'JetBrains Mono'", fontSize: 11 }} />
          <ReferenceLine y={baseP50} stroke={`${C.cyan}30`} strokeDasharray="4 4" />
          <Line type="monotone" dataKey="p50" stroke={C.cyan} strokeWidth={1.5} dot={false} isAnimationActive={false} />
          <Line type="monotone" dataKey="p95" stroke={C.amber} strokeWidth={1.5} dot={false} isAnimationActive={false} />
          <Line type="monotone" dataKey="p99" stroke={C.red} strokeWidth={1.5} dot={false} isAnimationActive={false} />
        </LineChart>
      </ResponsiveContainer>
      <div style={{ display: "flex", gap: 12, marginTop: 12 }}>
        {[
          { l: "p50", v: "~116ns", c: C.cyan },
          { l: "p95", v: "~142ns", c: C.amber },
          { l: "p99", v: "~165ns", c: C.red },
        ].map(({ l, v, c }) => (
          <div key={l} className="glass" style={{ flex: 1, padding: "10px 14px", borderColor: `${c}30` }}>
            <div className="mono" style={{ fontSize: 9, color: C.textMuted }}>{l}</div>
            <div className="mono" style={{ fontSize: 16, fontWeight: 700, color: c }}>{v}</div>
          </div>
        ))}
      </div>
    </div>
  );
}

// ─── CPU PROFILE FLAMEGRAPH ──────────────────────────────────────────────────
function FlameGraph() {
  const frames = [
    { label: "BenchmarkPressureHotKey", w: 1.0, d: 0, color: C.bg3 },
    { label: "runtime.goroutine", w: 0.92, d: 1, color: C.bg3 },
    { label: "dispatcher.Dispatch()", w: 0.88, d: 2, color: `${C.blue}90` },
    { label: "ringbuffer.Enqueue()", w: 0.55, d: 3, color: `${C.cyan}80` },
    { label: "runtime.lock2", w: 0.22, d: 3, color: `${C.red}80`, offset: 0.56 },
    { label: "worker.Process()", w: 0.42, d: 3, color: `${C.amber}80`, offset: 0.79 },
    { label: "atomic.CompareAndSwap", w: 0.48, d: 4, color: `${C.cyan}60` },
    { label: "pthread_cond_wait", w: 0.2, d: 4, color: `${C.red}70`, offset: 0.49 },
    { label: "runtime.usleep", w: 0.19, d: 4, color: `${C.red}60`, offset: 0.7 },
    { label: "aggregator.Batch()", w: 0.32, d: 4, color: `${C.green}70`, offset: 0.9 },
    { label: "memcpy / unsafe", w: 0.44, d: 5, color: `${C.cyan}50` },
    { label: "scheduler", w: 0.35, d: 5, color: `${C.red}50`, offset: 0.45 },
  ];

  const rowH = 28, pad = 2;
  const H = frames.reduce((mx, f) => Math.max(mx, f.d), 0) * (rowH + pad) + rowH + 60;

  return (
    <div className="glass" style={{ padding: 20 }}>
      <div className="mono" style={{ fontSize: 10, color: C.textMuted, marginBottom: 16, textTransform: "uppercase", letterSpacing: "0.08em" }}>cpu profile — call hierarchy</div>
      <div style={{ width: "100%", overflowX: "auto" }}>
        <svg viewBox={`0 0 820 ${H}`} style={{ width: "100%", minWidth: 400, display: "block" }}>
          {frames.map((f, i) => {
            const y = f.d * (rowH + pad) + 10;
            const x = (f.offset || 0) * 800 + 10;
            const w = f.w * 800;
            return (
              <g key={i}>
                <rect x={x} y={y} width={w} height={rowH} rx={3}
                  fill={f.color} stroke="rgba(0,0,0,0.3)" strokeWidth={0.5}
                  style={{ cursor: "pointer" }}
                />
                {w > 60 && (
                  <text x={x + 6} y={y + 18} fill="rgba(255,255,255,0.85)"
                    fontSize={10} fontFamily="'JetBrains Mono'" style={{ pointerEvents: "none" }}>
                    {w > 160 ? f.label : f.label.substring(0, Math.floor(w / 8))}
                  </text>
                )}
              </g>
            );
          })}
          <g>
            <rect x={10} y={H - 40} width={160} height={20} rx={3} fill={`${C.red}30`} stroke={`${C.red}50`} strokeWidth={0.5} />
            <text x={18} y={H - 26} fill={C.red} fontSize={10} fontFamily="'JetBrains Mono'">⚠ Scheduler bottleneck</text>
          </g>
          <g>
            <rect x={10} y={H - 16} width={320} height={14} rx={2} fill={`${C.cyan}10`} />
            <text x={18} y={H - 5} fill={C.textMuted} fontSize={9} fontFamily="'JetBrains Mono'">
              Runtime transitioned: algorithmic → scheduler-dominated at extreme throughput
            </text>
          </g>
        </svg>
      </div>
    </div>
  );
}

// ─── AGGREGATION LANE VIZ ─────────────────────────────────────────────────────
function AggregationLanes() {
  const [tick, setTick] = useState(0);
  useEffect(() => {
    const id = setInterval(() => setTick(t => t + 1), 150);
    return () => clearInterval(id);
  }, []);

  const workers = 4, lanes = 2;
  const packets = Array.from({ length: 6 }, (_, i) => ({
    id: i,
    progress: ((tick * 2 + i * 20) % 120) / 120,
    lane: i % lanes,
    worker: i % workers,
  }));

  return (
    <div className="glass" style={{ padding: 20 }}>
      <div className="mono" style={{ fontSize: 10, color: C.textMuted, marginBottom: 16, textTransform: "uppercase", letterSpacing: "0.08em" }}>aggregation pipeline — adaptive batching</div>
      <svg viewBox="0 0 640 200" style={{ width: "100%", display: "block" }}>
        {/* Workers */}
        {Array.from({ length: workers }, (_, i) => (
          <g key={i}>
            <rect x={10} y={20 + i * 42} width={80} height={30} rx={5}
              fill={C.bg3} stroke={`${C.amber}50`} strokeWidth={1} />
            <text x={50} y={39 + i * 42} textAnchor="middle"
              fill={C.amber} fontSize={9} fontFamily="'JetBrains Mono'">Worker {i}</text>
          </g>
        ))}

        {/* Aggregation lanes */}
        {Array.from({ length: lanes }, (_, i) => (
          <g key={i}>
            <rect x={230} y={40 + i * 80} width={140} height={50} rx={6}
              fill={C.bg2} stroke={`${C.green}40`} strokeWidth={1} />
            <text x={300} y={60 + i * 80} textAnchor="middle"
              fill={C.green} fontSize={9} fontFamily="'JetBrains Mono'" fontWeight={500}>Agg Lane {i}</text>
            <text x={300} y={75 + i * 80} textAnchor="middle"
              fill={C.textMuted} fontSize={8} fontFamily="'JetBrains Mono'">batch: {Math.round(8 + 6 * Math.sin(tick * 0.05 + i))}</text>
          </g>
        ))}

        {/* Redis */}
        <rect x={470} y={60} width={100} height={80} rx={6}
          fill={C.bg2} stroke={`${C.red}50`} strokeWidth={1} />
        <text x={520} y={95} textAnchor="middle" fill={C.red} fontSize={9} fontFamily="'JetBrains Mono'" fontWeight={500}>Redis</text>
        <text x={520} y={110} textAnchor="middle" fill={C.textMuted} fontSize={8} fontFamily="'JetBrains Mono'">Lua pipeline</text>
        <text x={520} y={125} textAnchor="middle" fill={C.textMuted} fontSize={8} fontFamily="'JetBrains Mono'">atomic exec</text>

        {/* Static flow lines */}
        {Array.from({ length: workers }, (_, wi) =>
          Array.from({ length: lanes }, (_, li) => (
            <line key={`${wi}-${li}`}
              x1={90} y1={35 + wi * 42}
              x2={230} y2={65 + li * 80}
              stroke="rgba(255,255,255,0.05)" strokeWidth={0.5}
            />
          ))
        )}
        <line x1={370} y1={90} x2={470} y2={100} stroke="rgba(255,255,255,0.08)" strokeWidth={1} />
        <line x1={370} y1={170} x2={470} y2={140} stroke="rgba(255,255,255,0.08)" strokeWidth={1} />

        {/* Animated packets */}
        {packets.map(p => {
          const startX = 90, startY = 35 + p.worker * 42;
          const midX = 230, midY = 65 + p.lane * 80;
          const endX = 470, endY = 100;
          let x, y;
          if (p.progress < 0.4) {
            const t = p.progress / 0.4;
            x = lerp(startX, midX, t);
            y = lerp(startY, midY, t);
          } else if (p.progress < 0.7) {
            x = midX + 70;
            y = midY + 25;
          } else {
            const t = (p.progress - 0.7) / 0.3;
            x = lerp(midX + 140, endX, t);
            y = lerp(midY + 25, endY, t);
          }
          return (
            <circle key={p.id} cx={x} cy={y} r={3}
              fill={C.cyan} opacity={0.8 - p.progress * 0.5}
              style={{ filter: `drop-shadow(0 0 4px ${C.cyan})` }}
            />
          );
        })}

        {/* Labels */}
        <text x={50} y={195} textAnchor="middle" fill={C.textMuted} fontSize={8} fontFamily="'JetBrains Mono'">Workers</text>
        <text x={300} y={195} textAnchor="middle" fill={C.textMuted} fontSize={8} fontFamily="'JetBrains Mono'">Aggregation lanes</text>
        <text x={520} y={155} textAnchor="middle" fill={C.textMuted} fontSize={8} fontFamily="'JetBrains Mono'">Datastore</text>
      </svg>
    </div>
  );
}

// ─── SOAK TEST SUMMARY ───────────────────────────────────────────────────────
function SoakTestResults() {
  const soakMetrics = [
    { label: "Total dispatched", value: "644M+", color: C.blue, icon: "→" },
    { label: "Completed reservations", value: "1.6M+", color: C.green, icon: "✓" },
    { label: "Rejected", value: "580M+", color: C.red, icon: "✗" },
    { label: "Throughput", value: "10.7M/s", color: C.cyan, icon: "⚡" },
  ];

  const soakLatency = [
    { label: "p50 residency", value: "~31s", note: "queue amplification", color: C.amber },
    { label: "p95 residency", value: "~56s", note: "near saturation limit", color: C.red },
    { label: "p99 residency", value: "~58s", note: "bounded by TTL", color: C.red },
  ];

  // Simulated 60s overload progression
  const soakHistory = Array.from({ length: 60 }, (_, i) => ({
    t: i,
    completed: Math.round(1600000 / 60 * (1 - Math.exp(-i * 0.08))),
    rejected: Math.round(580000000 / 60 * Math.min(1, i / 15)),
    queueDepth: Math.round(clamp(20 + i * 1.3, 20, 95)),
  }));

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 12 }}>
        {soakMetrics.map(m => (
          <div key={m.label} className="glass" style={{ padding: "16px 20px", borderTop: `1px solid ${m.color}40` }}>
            <div className="mono" style={{ fontSize: 9, color: C.textMuted, marginBottom: 6, textTransform: "uppercase", letterSpacing: "0.08em" }}>{m.label}</div>
            <div style={{ fontSize: 22, fontFamily: "'Syne'", fontWeight: 700, color: m.color }}>{m.value}</div>
          </div>
        ))}
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 16 }}>
        <div className="glass" style={{ padding: 20 }}>
          <div className="mono" style={{ fontSize: 10, color: C.textMuted, marginBottom: 12, textTransform: "uppercase", letterSpacing: "0.08em" }}>60s soak — queue saturation progression</div>
          <ResponsiveContainer width="100%" height={160}>
            <AreaChart data={soakHistory} margin={{ top: 5, right: 5, bottom: 0, left: 0 }}>
              <defs>
                <linearGradient id="qdGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stopColor={C.red} stopOpacity={0.5} />
                  <stop offset="100%" stopColor={C.red} stopOpacity={0.0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="2 4" stroke="rgba(255,255,255,0.03)" />
              <XAxis dataKey="t" tick={{ fill: C.textMuted, fontSize: 9, fontFamily: "'JetBrains Mono'" }} unit="s" />
              <YAxis tick={{ fill: C.textMuted, fontSize: 9, fontFamily: "'JetBrains Mono'" }} unit="%" />
              <Tooltip contentStyle={{ background: C.bg2, border: `1px solid ${C.border}`, borderRadius: 8, fontFamily: "'JetBrains Mono'", fontSize: 10 }} />
              <Area type="monotone" dataKey="queueDepth" name="Queue depth %" stroke={C.red} strokeWidth={2} fill="url(#qdGrad)" />
            </AreaChart>
          </ResponsiveContainer>
        </div>
        <div className="glass" style={{ padding: 20 }}>
          <div className="mono" style={{ fontSize: 10, color: C.textMuted, marginBottom: 12, textTransform: "uppercase", letterSpacing: "0.08em" }}>queue residency latency</div>
          {soakLatency.map(l => (
            <div key={l.label} style={{ marginBottom: 16 }}>
              <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 4 }}>
                <span className="mono" style={{ fontSize: 10, color: C.textMuted }}>{l.label}</span>
                <span className="mono" style={{ fontSize: 14, fontWeight: 700, color: l.color }}>{l.value}</span>
              </div>
              <div style={{ background: C.bg3, borderRadius: 3, height: 4, overflow: "hidden" }}>
                <div style={{
                  height: "100%", borderRadius: 3,
                  width: l.label === "p50" ? "53%" : l.label.includes("95") ? "95%" : "98%",
                  background: l.color, transition: "width 1s ease",
                }} />
              </div>
              <div className="mono" style={{ fontSize: 9, color: C.textMuted, marginTop: 3 }}>{l.note}</div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

// ─── MEMORY PROFILE ──────────────────────────────────────────────────────────
function MemoryProfile() {
  const allocs = [
    { label: "aggregation buffers", mb: 2.4, color: C.cyan },
    { label: "ring buffers (pre-alloc)", mb: 1.8, color: C.blue },
    { label: "telemetry structures", mb: 0.6, color: C.purple },
    { label: "goroutine stacks", mb: 0.4, color: C.amber },
    { label: "other", mb: 0.2, color: C.textDim },
  ];
  const total = allocs.reduce((s, a) => s + a.mb, 0);

  return (
    <div className="glass" style={{ padding: 20 }}>
      <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 16 }}>
        <div className="mono" style={{ fontSize: 10, color: C.textMuted, textTransform: "uppercase", letterSpacing: "0.08em" }}>memory residency — steady state</div>
        <Badge color={C.green}>0 allocs/op hotpath</Badge>
      </div>
      <div style={{ display: "flex", gap: 2, height: 24, borderRadius: 6, overflow: "hidden", marginBottom: 16 }}>
        {allocs.map(a => (
          <div key={a.label} style={{
            flex: a.mb / total,
            background: a.color,
            opacity: 0.85,
            transition: "flex 0.5s ease",
          }} />
        ))}
      </div>
      {allocs.map(a => (
        <div key={a.label} style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 8 }}>
          <div style={{ width: 10, height: 10, borderRadius: 2, background: a.color, flexShrink: 0 }} />
          <div style={{ flex: 1 }}>
            <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
              <span className="mono" style={{ fontSize: 10, color: C.text }}>{a.label}</span>
              <span className="mono" style={{ fontSize: 11, fontWeight: 500, color: a.color }}>{a.mb.toFixed(1)} MB</span>
            </div>
            <div style={{ background: C.bg3, borderRadius: 2, height: 3, marginTop: 3 }}>
              <div style={{ width: `${(a.mb / total) * 100}%`, height: "100%", background: a.color, borderRadius: 2 }} />
            </div>
          </div>
        </div>
      ))}
      <div style={{ marginTop: 16, padding: "12px 16px", background: C.bg3, borderRadius: 8, borderLeft: `3px solid ${C.cyan}` }}>
        <div className="mono" style={{ fontSize: 10, color: C.cyan, fontWeight: 500, marginBottom: 4 }}>Front-loaded allocation strategy</div>
        <div className="mono" style={{ fontSize: 9, color: C.textMuted, lineHeight: 1.6 }}>
          Runtime intentionally pre-allocates all buffers at startup. Hotpath execution achieves<br />
          0 B/op, 0 allocs/op under sustained load — verified across all benchmark iterations.
        </div>
      </div>
    </div>
  );
}

// ─── ENGINEERING STAGE ────────────────────────────────────────────────────────
function StageCard({ num, title, status, items, metric, metricLabel, color = C.cyan }) {
  return (
    <div className="glass" style={{ padding: 24, position: "relative", overflow: "hidden" }}>
      <div style={{
        position: "absolute", top: 0, left: 0, bottom: 0, width: 3,
        background: `linear-gradient(180deg, ${color}, ${color}30)`,
      }} />
      <div style={{ display: "flex", gap: 12, alignItems: "flex-start" }}>
        <div className="syne" style={{
          width: 36, height: 36, borderRadius: 8, flexShrink: 0,
          background: `${color}15`, border: `1px solid ${color}40`,
          display: "flex", alignItems: "center", justifyContent: "center",
          fontSize: 16, fontWeight: 800, color,
        }}>{num}</div>
        <div style={{ flex: 1 }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 8 }}>
            <div style={{ fontSize: 14, fontWeight: 600, fontFamily: "'Syne'", color }}>{title}</div>
            <Badge color={color}>{status}</Badge>
          </div>
          <ul style={{ listStyle: "none", display: "flex", flexDirection: "column", gap: 4 }}>
            {items.map((item, i) => (
              <li key={i} style={{ display: "flex", alignItems: "center", gap: 6 }}>
                <div style={{ width: 4, height: 4, borderRadius: "50%", background: color, opacity: 0.6, flexShrink: 0 }} />
                <span className="mono" style={{ fontSize: 10, color: C.textMuted }}>{item}</span>
              </li>
            ))}
          </ul>
          {metric && (
            <div style={{ marginTop: 12, padding: "8px 12px", background: `${color}08`, borderRadius: 6 }}>
              <span className="mono" style={{ fontSize: 11, fontWeight: 700, color }}>{metric}</span>
              <span className="mono" style={{ fontSize: 10, color: C.textMuted, marginLeft: 8 }}>{metricLabel}</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── NAV ──────────────────────────────────────────────────────────────────────
function Nav({ active, setActive }) {
  const items = [
    { id: "overview", label: "Overview" },
    { id: "architecture", label: "Architecture" },
    { id: "evolution", label: "Evolution" },
    { id: "benchmarks", label: "Benchmarks" },
    { id: "telemetry", label: "Telemetry" },
    { id: "profiling", label: "Profiling" },
  ];
  return (
    <nav style={{
      position: "sticky", top: 0, zIndex: 100,
      background: "rgba(8,9,12,0.9)", backdropFilter: "blur(16px)",
      borderBottom: `1px solid ${C.border}`,
      padding: "0 40px",
    }}>
      <div style={{ maxWidth: 1100, margin: "0 auto", display: "flex", alignItems: "center", gap: 0, height: 52 }}>
        <div className="syne" style={{ fontSize: 15, fontWeight: 800, color: C.cyan, marginRight: 32, letterSpacing: "-0.01em" }}>
          Flux<span style={{ color: C.text, opacity: 0.6 }}>Runtime</span>
        </div>
        {items.map(item => (
          <button key={item.id} onClick={() => {
            setActive(item.id);
            document.getElementById(item.id)?.scrollIntoView({ behavior: "smooth" });
          }} style={{
            background: "none", border: "none", cursor: "pointer",
            padding: "8px 14px",
            fontSize: 12, fontFamily: "'JetBrains Mono'", letterSpacing: "0.02em",
            color: active === item.id ? C.cyan : C.textMuted,
            borderBottom: active === item.id ? `2px solid ${C.cyan}` : "2px solid transparent",
            transition: "all 0.2s",
          }}>{item.label}</button>
        ))}
        <div style={{ marginLeft: "auto" }}>
          <Badge color={C.green}>Go · ARM64 · Darwin</Badge>
        </div>
      </div>
    </nav>
  );
}

// ─── HERO ─────────────────────────────────────────────────────────────────────
function Hero() {
  return (
    <section id="overview" style={{ padding: "100px 40px 80px", position: "relative", overflow: "hidden" }}>
      {/* Bg grid */}
      <div style={{
        position: "absolute", inset: 0,
        backgroundImage: `
          linear-gradient(${C.border} 1px, transparent 1px),
          linear-gradient(90deg, ${C.border} 1px, transparent 1px)
        `,
        backgroundSize: "64px 64px",
        maskImage: "radial-gradient(ellipse 80% 70% at 50% 0%, black 30%, transparent 100%)",
      }} />
      {/* Glow */}
      <div style={{
        position: "absolute", top: 40, left: "50%", transform: "translateX(-50%)",
        width: 600, height: 300,
        background: `radial-gradient(ellipse, ${C.cyanGlow} 0%, transparent 70%)`,
        pointerEvents: "none",
      }} />

      <div style={{ maxWidth: 1100, margin: "0 auto", position: "relative" }}>
        <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 24 }}>
          <Badge color={C.cyan}>Research Runtime</Badge>
          <Badge color={C.purple}>High Performance</Badge>
          <Badge color={C.amber}>Go · Apple M1</Badge>
        </div>

        <h1 className="syne" style={{
          fontSize: "clamp(40px, 6vw, 72px)", fontWeight: 800,
          lineHeight: 1.05, letterSpacing: "-0.03em",
          marginBottom: 20,
        }}>
          <span style={{ color: C.cyan }}>FluxRuntime</span>
          <br />
          <span style={{ color: C.text, opacity: 0.85 }}>Reservation Execution</span>
          <br />
          <span style={{ color: C.text, opacity: 0.5 }}>at Extreme Throughput</span>
        </h1>

        <p style={{ fontSize: 16, color: C.textMuted, maxWidth: 560, lineHeight: 1.7, marginBottom: 40, fontWeight: 300 }}>
          A research-oriented high-performance runtime exploring deterministic shard routing, lock-free ring buffers, adaptive batching, and probabilistic overload shedding — benchmarked at 10.5M req/s with zero allocations on the hotpath.
        </p>

        <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 12, maxWidth: 680 }}>
          {[
            { v: "10.5M", u: "req/s", l: "Peak throughput" },
            { v: "115.8", u: "ns/op", l: "Hotkey benchmark" },
            { v: "0", u: "allocs/op", l: "Steady state" },
            { v: "644M+", u: "dispatched", l: "60s soak test" },
          ].map(m => (
            <div key={m.l} className="glass" style={{ padding: "16px 18px", borderTop: `1px solid ${C.cyanDim}` }}>
              <div className="syne" style={{ fontSize: 24, fontWeight: 800, color: C.cyan, lineHeight: 1 }}>{m.v}</div>
              <div className="mono" style={{ fontSize: 10, color: C.textMuted, marginTop: 2 }}>{m.u}</div>
              <div className="mono" style={{ fontSize: 9, color: C.textDim, marginTop: 4 }}>{m.l}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

// ─── BENCHMARK TABLE ─────────────────────────────────────────────────────────
function BenchmarkTable() {
  const rows = [
    { workers: 2, throughput: "5.5M", ns: "211.3", allocs: 0, bytes: 0 },
    { workers: 4, throughput: "6.4M", ns: "178.0", allocs: 0, bytes: 0 },
    { workers: 8, throughput: "10.3M", ns: "116.3", allocs: 0, bytes: 0 },
    { workers: 16, throughput: "11.4M", ns: "106.2", allocs: 0, bytes: 0 },
    { workers: "8 (hotkey)", throughput: "10.5M", ns: "115.8", allocs: 0, bytes: 0, highlight: true },
  ];

  return (
    <div className="glass" style={{ overflow: "hidden" }}>
      <div style={{ borderBottom: `1px solid ${C.border}`, padding: "16px 20px" }}>
        <div className="mono" style={{ fontSize: 10, color: C.textMuted, textTransform: "uppercase", letterSpacing: "0.08em" }}>benchmark — BenchmarkPressureHotKey · Apple M1 · Darwin ARM64</div>
      </div>
      <div style={{ overflowX: "auto" }}>
        <table style={{ width: "100%", borderCollapse: "collapse" }}>
          <thead>
            <tr style={{ borderBottom: `1px solid ${C.border}` }}>
              {["Workers", "Throughput", "ns/op", "allocs/op", "B/op"].map(h => (
                <th key={h} style={{ padding: "10px 20px", textAlign: "left", fontFamily: "'JetBrains Mono'", fontSize: 9, color: C.textMuted, fontWeight: 500, textTransform: "uppercase", letterSpacing: "0.08em" }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((r, i) => (
              <tr key={i} style={{
                borderBottom: `1px solid ${C.border}`,
                background: r.highlight ? C.cyanGlow : "transparent",
              }}>
                <td style={{ padding: "12px 20px", fontFamily: "'JetBrains Mono'", fontSize: 12, color: r.highlight ? C.cyan : C.text }}>{r.workers}</td>
                <td style={{ padding: "12px 20px", fontFamily: "'JetBrains Mono'", fontSize: 12, fontWeight: 700, color: r.highlight ? C.cyan : C.green }}>{r.throughput}</td>
                <td style={{ padding: "12px 20px", fontFamily: "'JetBrains Mono'", fontSize: 12, color: r.highlight ? C.amber : C.text }}>{r.ns}</td>
                <td style={{ padding: "12px 20px", fontFamily: "'JetBrains Mono'", fontSize: 12, color: C.green }}>0</td>
                <td style={{ padding: "12px 20px", fontFamily: "'JetBrains Mono'", fontSize: 12, color: C.green }}>0</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// ─── TELEMETRY DASHBOARD ──────────────────────────────────────────────────────
function TelemetryDashboard() {
  const [data, setData] = useState(() =>
    Array.from({ length: 30 }, (_, i) => ({
      t: i,
      throughput: 9.5 + Math.random() * 1.5,
      queueDepth: 30 + Math.random() * 40,
      rejectionRate: 5 + Math.random() * 15,
      batchSize: 12 + Math.random() * 8,
    }))
  );

  useEffect(() => {
    const id = setInterval(() => {
      setData(prev => {
        const last = prev[prev.length - 1];
        const next = {
          t: last.t + 1,
          throughput: clamp(last.throughput + (Math.random() - 0.5) * 0.4, 8.5, 11.5),
          queueDepth: clamp(last.queueDepth + (Math.random() - 0.5) * 8, 10, 95),
          rejectionRate: clamp(last.rejectionRate + (Math.random() - 0.5) * 4, 2, 45),
          batchSize: clamp(last.batchSize + (Math.random() - 0.5) * 2, 8, 24),
        };
        return [...prev.slice(-30), next];
      });
    }, 800);
    return () => clearInterval(id);
  }, []);

  const latest = data[data.length - 1];

  return (
    <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
      {/* Live counters */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: 12 }}>
        {[
          { l: "Throughput", v: latest.throughput.toFixed(1), u: "M req/s", c: C.cyan },
          { l: "Queue depth", v: Math.round(latest.queueDepth), u: "%", c: latest.queueDepth > 70 ? C.red : C.amber },
          { l: "Rejection rate", v: latest.rejectionRate.toFixed(1), u: "%", c: C.red },
          { l: "Batch size", v: latest.batchSize.toFixed(0), u: "req/batch", c: C.green },
        ].map(({ l, v, u, c }) => (
          <div key={l} className="glass" style={{ padding: "16px 20px", borderTop: `1px solid ${c}40` }}>
            <div className="mono" style={{ fontSize: 9, color: C.textMuted, textTransform: "uppercase", letterSpacing: "0.08em", marginBottom: 6 }}>{l}</div>
            <div style={{ display: "flex", alignItems: "baseline", gap: 4 }}>
              <span className="syne" style={{ fontSize: 24, fontWeight: 700, color: c }}>{v}</span>
              <span className="mono" style={{ fontSize: 10, color: C.textMuted }}>{u}</span>
            </div>
            <div style={{ width: 6, height: 6, borderRadius: "50%", background: c, marginTop: 8, animation: "glow-pulse 1.2s ease-in-out infinite" }} />
          </div>
        ))}
      </div>

      {/* Multi-metric chart */}
      <div className="glass" style={{ padding: 20 }}>
        <div className="mono" style={{ fontSize: 10, color: C.textMuted, marginBottom: 16, textTransform: "uppercase", letterSpacing: "0.08em" }}>runtime telemetry — live feed (30s window)</div>
        <ResponsiveContainer width="100%" height={200}>
          <LineChart data={data} margin={{ top: 5, right: 10, bottom: 5, left: 0 }}>
            <CartesianGrid strokeDasharray="2 4" stroke="rgba(255,255,255,0.04)" />
            <XAxis dataKey="t" hide />
            <YAxis yAxisId="left" tick={{ fill: C.textMuted, fontSize: 9, fontFamily: "'JetBrains Mono'" }} />
            <YAxis yAxisId="right" orientation="right" tick={{ fill: C.textMuted, fontSize: 9, fontFamily: "'JetBrains Mono'" }} />
            <Tooltip contentStyle={{ background: C.bg2, border: `1px solid ${C.border}`, borderRadius: 8, fontFamily: "'JetBrains Mono'", fontSize: 10 }} />
            <Line yAxisId="left" type="monotone" dataKey="queueDepth" name="Queue %" stroke={C.blue} strokeWidth={1.5} dot={false} isAnimationActive={false} />
            <Line yAxisId="left" type="monotone" dataKey="rejectionRate" name="Rejection %" stroke={C.red} strokeWidth={1.5} dot={false} isAnimationActive={false} />
            <Line yAxisId="right" type="monotone" dataKey="batchSize" name="Batch size" stroke={C.green} strokeWidth={1.5} dot={false} strokeDasharray="4 2" isAnimationActive={false} />
          </LineChart>
        </ResponsiveContainer>
        <div style={{ display: "flex", gap: 16, marginTop: 8 }}>
          {[
            { l: "Queue depth %", c: C.blue },
            { l: "Rejection %", c: C.red },
            { l: "Batch size (dashed)", c: C.green },
          ].map(({ l, c }) => (
            <div key={l} style={{ display: "flex", alignItems: "center", gap: 4 }}>
              <div style={{ width: 16, height: 2, background: c }} />
              <span className="mono" style={{ fontSize: 9, color: C.textMuted }}>{l}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

// ─── MAIN APP ─────────────────────────────────────────────────────────────────
export default function App() {
  const [navActive, setNavActive] = useState("overview");

  useEffect(() => {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach(e => { if (e.isIntersecting) setNavActive(e.target.id); });
      },
      { threshold: 0.3 }
    );
    ["overview", "architecture", "evolution", "benchmarks", "telemetry", "profiling"].forEach(id => {
      const el = document.getElementById(id);
      if (el) observer.observe(el);
    });
    return () => observer.disconnect();
  }, []);

  return (
    <div style={{ background: C.bg, minHeight: "100vh", color: C.text }}>
      <style>{globalStyle}</style>

      <Nav active={navActive} setActive={setNavActive} />

      <Hero />

      {/* ── ARCHITECTURE ── */}
      <Section id="architecture" style={{ padding: "80px 40px" }}>
        <div style={{ maxWidth: 1100, margin: "0 auto" }}>
          <div style={{ marginBottom: 40 }}>
            <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 12 }}>
              <Badge color={C.blue}>System Design</Badge>
            </div>
            <h2 className="syne" style={{ fontSize: 36, fontWeight: 800, letterSpacing: "-0.02em", marginBottom: 12 }}>
              Core <span style={{ color: C.blue }}>Architecture</span>
            </h2>
            <p style={{ color: C.textMuted, fontSize: 14, maxWidth: 520, lineHeight: 1.7 }}>
              Deterministic FNV-1a routing feeds a lock-free ring buffer array. Parallel worker goroutines drain into adaptive aggregation lanes, compressing datastore write amplification through Lua pipeline execution.
            </p>
          </div>

          <ArchDiagram />

          <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 16, marginTop: 24 }}>
            {[
              { title: "Dispatcher", sub: "FNV-1a hashing ensures stable shard ownership — identical keys always route to the same ring buffer, eliminating cross-shard contention.", color: C.blue, icon: "⇢" },
              { title: "Lock-Free RingBuffer", sub: "Bounded SPSC/MPMC queue using atomic CAS operations. Zero heap allocations on the enqueue/dequeue hotpath via pre-allocated slot arrays.", color: C.cyan, icon: "○" },
              { title: "Aggregation Lanes", sub: "Microbatch construction amortizes per-request datastore round-trips. Adaptive sizing responds to downstream saturation pressure.", color: C.green, icon: "≡" },
            ].map(c => (
              <div key={c.title} className="glass" style={{ padding: 20 }}>
                <div style={{ fontSize: 20, marginBottom: 8 }}>{c.icon}</div>
                <div className="syne" style={{ fontSize: 14, fontWeight: 700, color: c.color, marginBottom: 6 }}>{c.title}</div>
                <div className="mono" style={{ fontSize: 10, color: C.textMuted, lineHeight: 1.6 }}>{c.sub}</div>
              </div>
            ))}
          </div>

          {/* Ring Buffer Animation */}
          <div style={{ marginTop: 24 }}>
            <div className="glass" style={{ padding: 24 }}>
              <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 20 }}>
                <div>
                  <div className="mono" style={{ fontSize: 10, color: C.textMuted, textTransform: "uppercase", letterSpacing: "0.08em", marginBottom: 4 }}>lock-free ring buffer — live visualization</div>
                  <div className="mono" style={{ fontSize: 9, color: C.textDim }}>16-slot bounded queue · atomic head/tail · zero-allocation enqueue</div>
                </div>
                <Badge color={C.cyan}>0 allocs/op</Badge>
              </div>
              <RingBufferViz />
            </div>
          </div>
        </div>
      </Section>

      {/* ── EVOLUTION ── */}
      <Section id="evolution" style={{ padding: "80px 40px", background: `linear-gradient(180deg, ${C.bg} 0%, ${C.bg1} 100%)` }}>
        <div style={{ maxWidth: 1100, margin: "0 auto" }}>
          <div style={{ marginBottom: 40 }}>
            <Badge color={C.purple}>Engineering Journey</Badge>
            <h2 className="syne" style={{ fontSize: 36, fontWeight: 800, letterSpacing: "-0.02em", margin: "12px 0" }}>
              Implementation <span style={{ color: C.purple }}>Evolution</span>
            </h2>
            <p style={{ color: C.textMuted, fontSize: 14, maxWidth: 520, lineHeight: 1.7 }}>
              Five progressive engineering stages — from baseline contention-prone implementation to a scheduler-bound, allocation-free runtime operating at 10.5M req/s.
            </p>
          </div>

          <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
            <StageCard num={1} title="Baseline Runtime" status="discovered" color={C.red}
              items={[
                "Initial shard queues with mutex-based synchronization",
                "Goroutine-per-worker model with channel communication",
                "FNV-1a deterministic routing established",
                "Discovered: lock contention, saturation instability, blocking enqueue",
              ]}
              metric="~3.2M req/s" metricLabel="initial baseline" />
            <StageCard num={2} title="Lock-Free Evolution" status="implemented" color={C.amber}
              items={[
                "Replaced mutex queues with atomic CAS ring buffers",
                "Bounded queue size prevents unbounded memory growth",
                "Zero-allocation enqueue on steady-state hotpath",
                "Major scheduler contention reduction observed",
              ]}
              metric="~7.8M req/s" metricLabel="+144% improvement" />
            <StageCard num={3} title="Aggregation Pipelines" status="optimized" color={C.cyan}
              items={[
                "Implemented adaptive aggregation lanes between workers and datastore",
                "Dynamic batch sizing responds to downstream pressure",
                "Redis Lua atomic reservation pipeline reduces RTT amplification",
                "Queue amortization enables sustained overload handling",
              ]}
              metric="~9.5M req/s" metricLabel="aggregation gain" />
            <StageCard num={4} title="Runtime Telemetry" status="instrumented" color={C.blue}
              items={[
                "p50/p95/p99 latency histogram tracking via atomic counters",
                "Queue depth, rejection rate, batch evolution metrics",
                "Throughput counters with nanosecond resolution",
                "Zero telemetry overhead on critical path via background aggregation",
              ]}
              metric="0ns overhead" metricLabel="telemetry cost" />
            <StageCard num={5} title="Overload Control" status="stabilized" color={C.green}
              items={[
                "Probabilistic early rejection under queue pressure",
                "Adaptive shedding curves calibrated to buffer fill ratio",
                "Queue stabilization under 60s sustained overload (soak test)",
                "Bounded memory under extreme load via rejection backpressure",
              ]}
              metric="10.5M req/s" metricLabel="final benchmark" />
          </div>

          {/* Aggregation animation */}
          <div style={{ marginTop: 24 }}>
            <AggregationLanes />
          </div>
        </div>
      </Section>

      {/* ── BENCHMARKS ── */}
      <Section id="benchmarks" style={{ padding: "80px 40px" }}>
        <div style={{ maxWidth: 1100, margin: "0 auto" }}>
          <div style={{ marginBottom: 40 }}>
            <Badge color={C.amber}>Performance</Badge>
            <h2 className="syne" style={{ fontSize: 36, fontWeight: 800, letterSpacing: "-0.02em", margin: "12px 0" }}>
              Benchmark <span style={{ color: C.amber }}>Results</span>
            </h2>
          </div>

          <div className="glass" style={{ padding: "16px 20px", marginBottom: 16, fontFamily: "'JetBrains Mono'", fontSize: 11, borderLeft: `3px solid ${C.cyan}` }}>
            <div style={{ color: C.textMuted, fontSize: 9, marginBottom: 4 }}>go test -bench=BenchmarkPressureHotKey -benchmem</div>
            <div style={{ color: C.cyan }}>BenchmarkPressureHotKey-8 <span style={{ color: C.text }}>   10509103    </span><span style={{ color: C.amber }}>115.8 ns/op</span><span style={{ color: C.green }}>   0 B/op   0 allocs/op</span></div>
          </div>

          <BenchmarkTable />

          <div style={{ marginTop: 24 }}>
            <ScalingChart />
          </div>

          <div style={{ marginTop: 24 }}>
            <div style={{ marginBottom: 24 }}>
              <div className="syne" style={{ fontSize: 20, fontWeight: 700, marginBottom: 8 }}>Soak Test — 60s Sustained Overload</div>
              <div className="mono" style={{ fontSize: 11, color: C.textMuted, lineHeight: 1.6, maxWidth: 600 }}>
                60-second execution under continuous write pressure. Queue saturation stabilized by adaptive rejection with bounded memory preservation throughout.
              </div>
            </div>
            <SoakTestResults />
          </div>
        </div>
      </Section>

      {/* ── TELEMETRY ── */}
      <Section id="telemetry" style={{ padding: "80px 40px", background: `linear-gradient(180deg, ${C.bg} 0%, ${C.bg1} 100%)` }}>
        <div style={{ maxWidth: 1100, margin: "0 auto" }}>
          <div style={{ marginBottom: 40 }}>
            <Badge color={C.green}>Live Simulation</Badge>
            <h2 className="syne" style={{ fontSize: 36, fontWeight: 800, letterSpacing: "-0.02em", margin: "12px 0" }}>
              Runtime <span style={{ color: C.green }}>Telemetry</span>
            </h2>
            <p style={{ color: C.textMuted, fontSize: 14, maxWidth: 520, lineHeight: 1.7 }}>
              Animated telemetry simulation reflecting real benchmark characteristics. Queue depth, rejection rates, batch evolution, and latency percentiles updated in real time.
            </p>
          </div>

          <TelemetryDashboard />

          <div style={{ marginTop: 24 }}>
            <LatencyChart />
          </div>

          <div style={{ marginTop: 24 }}>
            <QueueSaturation />
          </div>
        </div>
      </Section>

      {/* ── PROFILING ── */}
      <Section id="profiling" style={{ padding: "80px 40px" }}>
        <div style={{ maxWidth: 1100, margin: "0 auto" }}>
          <div style={{ marginBottom: 40 }}>
            <Badge color={C.red}>CPU + Memory</Badge>
            <h2 className="syne" style={{ fontSize: 36, fontWeight: 800, letterSpacing: "-0.02em", margin: "12px 0" }}>
              Profiling <span style={{ color: C.red }}>Analysis</span>
            </h2>
          </div>

          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 24, marginBottom: 24 }}>
            <div>
              <div className="syne" style={{ fontSize: 18, fontWeight: 700, marginBottom: 16 }}>CPU Hotspots</div>
              <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                {[
                  { fn: "runtime.usleep", pct: 34, c: C.red, note: "OS thread sleep — scheduler bottleneck" },
                  { fn: "pthread_cond_wait", pct: 28, c: C.red, note: "kernel mutex wait — goroutine parking" },
                  { fn: "runtime.lock2", pct: 18, c: C.amber, note: "internal Go runtime locking" },
                  { fn: "atomic.CompareAndSwap", pct: 11, c: C.cyan, note: "ring buffer CAS operations" },
                  { fn: "aggregator.Batch()", pct: 6, c: C.green, note: "batch construction overhead" },
                  { fn: "other", pct: 3, c: C.textDim, note: "dispatch, hash, misc" },
                ].map(({ fn, pct, c, note }) => (
                  <div key={fn}>
                    <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 3 }}>
                      <span className="mono" style={{ fontSize: 10, color: c }}>{fn}</span>
                      <span className="mono" style={{ fontSize: 10, color: C.textMuted }}>{pct}%</span>
                    </div>
                    <div style={{ background: C.bg3, borderRadius: 2, height: 5, marginBottom: 3 }}>
                      <div style={{ width: `${pct}%`, height: "100%", background: c, borderRadius: 2 }} />
                    </div>
                    <div className="mono" style={{ fontSize: 9, color: C.textDim }}>{note}</div>
                  </div>
                ))}
              </div>
            </div>

            <div>
              <div className="syne" style={{ fontSize: 18, fontWeight: 700, marginBottom: 16 }}>Profiling Findings</div>
              <div style={{ display: "flex", flexDirection: "column", gap: 12 }}>
                {[
                  { title: "Scheduler-dominated regime", body: "At extreme throughput, the runtime transitioned from algorithmic bottlenecks (CAS loops, hashing) into scheduler-dominated bottlenecks (usleep, cond_wait). This indicates the workload saturated available CPU scheduling capacity on 8 cores.", color: C.red },
                  { title: "Lock-free success confirmed", body: "CAS operations represent only 11% of CPU time — a significant reduction from the mutex-dominated baseline. The lock-free ring buffer implementation successfully eliminated the primary algorithmic bottleneck.", color: C.cyan },
                  { title: "Aggregation amortization", body: "Aggregator batch construction at 6% CPU confirms the pipeline compression is functioning: many worker writes are amortized into fewer datastore operations with minimal per-operation overhead.", color: C.green },
                ].map(({ title, body, color }) => (
                  <div key={title} className="glass" style={{ padding: 16, borderLeft: `3px solid ${color}60` }}>
                    <div className="mono" style={{ fontSize: 11, fontWeight: 600, color, marginBottom: 6 }}>{title}</div>
                    <div className="mono" style={{ fontSize: 10, color: C.textMuted, lineHeight: 1.6 }}>{body}</div>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <FlameGraph />

          <div style={{ marginTop: 24 }}>
            <div className="syne" style={{ fontSize: 18, fontWeight: 700, marginBottom: 16 }}>Memory Profile</div>
            <MemoryProfile />
          </div>
        </div>
      </Section>

      {/* ── FOOTER ── */}
      <footer style={{
        borderTop: `1px solid ${C.border}`,
        padding: "40px 40px",
        background: C.bg1,
      }}>
        <div style={{ maxWidth: 1100, margin: "0 auto", display: "flex", justifyContent: "space-between", alignItems: "center", flexWrap: "wrap", gap: 16 }}>
          <div>
            <div className="syne" style={{ fontSize: 16, fontWeight: 800, color: C.cyan, marginBottom: 4 }}>FluxRuntime</div>
            <div className="mono" style={{ fontSize: 10, color: C.textMuted }}>Research-oriented high-performance reservation execution runtime · Go · Apple M1 · ARM64</div>
          </div>
          <div style={{ display: "flex", gap: 8 }}>
            <Badge color={C.cyan}>10.5M req/s</Badge>
            <Badge color={C.green}>0 allocs/op</Badge>
            <Badge color={C.amber}>115.8 ns/op</Badge>
          </div>
        </div>
      </footer>
    </div>
  );
}
