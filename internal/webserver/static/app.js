'use strict';
const $ = (id) => document.getElementById(id);
const I18N = {
  ru: {
    tab_scan:'Сканирование', tab_tools:'Инструменты', tab_history:'История', tab_account:'Аккаунт', logout:'Выйти',
    cancel:'Остановить', scan_title:'Сканирование адресов WARP', preset:'Пресет', protocol:'Протокол', expert:'Экспертный режим',
    preset_standard:'Standard', preset_durable:'Durable', preset_full:'Full',
    hint_standard:'Быстрая выборка: 5 адресов на подсеть.', hint_durable:'Выборка + 10 пингов внутри туннеля для проверки стабильности.', hint_full:'Все адреса и 10 пингов в туннеле. Долго.',
    inner_protocol:'Внутренний протокол', through:'Через endpoint (WARP-in-WARP)', timeout:'Таймаут, с', jobs:'Потоков', tunnel_pings:'Пингов в туннеле',
    custom_target:'Свои диапазоны', ping_target:'Цель пинга', include_countries:'Только страны', include_nodes:'Только узлы', best_by:'Лучший по', sweep_ports:'Перебор портов', speed_test:'Проверка скорости',
    scan_start:'Начать сканирование', results:'Результаты', text_report:'Текстовый отчёт', col_endpoint:'Адрес', col_status:'Статус', col_place:'Место',
    tool_junk:'Поиск AWG junk / I1', tool_junk_hint:'Подбирает параметры junk-пакетов и I1 для AmneziaWG.', run:'Запустить',
    tool_sni:'Поиск MASQUE SNI', tool_sni_hint:'Подбирает SNI, с которым MASQUE проходит в вашей сети.',
    socks_hint:'SOCKS слушает только 127.0.0.1 на сервере. Снаружи — через SSH-туннель: ssh -L 1080:127.0.0.1:1080 user@сервер.', socks_start:'Запустить SOCKS',
    clear:'Очистить', history_empty:'История пуста.', reg_fresh:'Новый аккаунт', reg_timeout:'Таймаут, с', register:'Зарегистрировать аккаунт', export:'Экспорт', delete:'Удалить', import_account:'Импорт аккаунта (JSON)', import:'Импортировать',
    login:'Войти', login_hint:'Введите токен доступа (файл token в каталоге данных сервера).',
    copy:'Копировать', download:'Скачать', open:'Открыть', config:'Конфиг',
    acc_yes:'WARP-аккаунт сохранён на сервере.', acc_no:'Аккаунта нет. Зарегистрируйте или импортируйте его.',
    working:'работает', torn:'не держит', failed:'нет', running:'Выполняется', done:'Готово', canceled:'Остановлено', busy:'Уже выполняется другая операция.',
    confirm_delete_acc:'Удалить аккаунт с сервера?', confirm_clear:'Очистить всю историю?', copied:'Скопировано', need_account:'Сначала зарегистрируйте или импортируйте аккаунт.',
    found:'Найдено рабочих', of:'из', best:'лучший', apply:'Применить параметры', junk_done:'Подобрано', sni_done:'Подобрано',
    tg_check:'Проверить Telegram', tg_only:'Только с Telegram', tg_hint:'Telegram: через каждый туннель отправляется настоящий MTProto-запрос ко всем 5 дата-центрам. Доступным считается только адрес, где ответили все пять; такие адреса идут первыми.', tg_ok:'Telegram работает', tg_blocked:'заблокирован', exclude_countries:'Исключить страны', exclude_nodes:'Исключить узлы', flags:'Флаги', copy_best:'Копировать лучший', csv:'CSV', apply_to_scan:'Применить к сканированию', saved_at:'Последний результат', applied:'Параметры применены, режим эксперта включён', confirm_token:'Введите токен доступа для экспорта аккаунта:'
  },
  en: {
    tab_scan:'Scan', tab_tools:'Tools', tab_history:'History', tab_account:'Account', logout:'Log out',
    cancel:'Stop', scan_title:'Scan WARP endpoints', preset:'Preset', protocol:'Protocol', expert:'Expert mode',
    preset_standard:'Standard', preset_durable:'Durable', preset_full:'Full',
    hint_standard:'Quick sample: 5 addresses per subnet.', hint_durable:'Sample + 10 in-tunnel pings to check stability.', hint_full:'All addresses and 10 in-tunnel pings. Slow.',
    inner_protocol:'Inner protocol', through:'Through endpoint (WARP-in-WARP)', timeout:'Timeout, s', jobs:'Workers', tunnel_pings:'Tunnel pings',
    custom_target:'Custom ranges', ping_target:'Ping target', include_countries:'Only countries', include_nodes:'Only nodes', best_by:'Best by', sweep_ports:'Port sweep', speed_test:'Speed test',
    scan_start:'Start scan', results:'Results', text_report:'Text report', col_endpoint:'Endpoint', col_status:'Status', col_place:'Location',
    tool_junk:'AWG junk / I1 finder', tool_junk_hint:'Finds junk packet parameters and I1 for AmneziaWG.', run:'Run',
    tool_sni:'MASQUE SNI finder', tool_sni_hint:'Finds an SNI that MASQUE passes with on your network.',
    socks_hint:'SOCKS listens on 127.0.0.1 of the server only. Reach it via SSH: ssh -L 1080:127.0.0.1:1080 user@server.', socks_start:'Start SOCKS',
    clear:'Clear', history_empty:'No history yet.', reg_fresh:'Fresh account', reg_timeout:'Timeout, s', register:'Register account', export:'Export', delete:'Delete', import_account:'Import account (JSON)', import:'Import',
    login:'Log in', login_hint:'Enter the access token (the "token" file in the server data directory).',
    copy:'Copy', download:'Download', open:'Open', config:'Config',
    acc_yes:'A WARP account is stored on the server.', acc_no:'No account yet. Register or import one.',
    working:'working', torn:'unstable', failed:'failed', running:'Running', done:'Done', canceled:'Stopped', busy:'Another operation is already running.',
    confirm_delete_acc:'Delete the account from the server?', confirm_clear:'Clear all history?', copied:'Copied', need_account:'Register or import an account first.',
    found:'Working', of:'of', best:'best', apply:'Apply options', junk_done:'Found', sni_done:'Found',
    tg_check:'Check Telegram', tg_only:'Only with Telegram', tg_hint:'Telegram: a real MTProto request is sent to all 5 data centres through every tunnel. An endpoint counts only when all five answer; those are ranked first.', tg_ok:'Telegram works', tg_blocked:'blocked', exclude_countries:'Exclude countries', exclude_nodes:'Exclude nodes', flags:'Flags', copy_best:'Copy best', csv:'CSV', apply_to_scan:'Apply to scan', saved_at:'Latest result', applied:'Options applied, expert mode enabled', confirm_token:'Enter the access token to export the account:'
  }
};
let lang = (() => { try { return localStorage.getItem('lang') || (navigator.language || 'ru').slice(0,2); } catch { return 'ru'; } })();
if (!I18N[lang]) lang = 'en';
const t = (k) => (I18N[lang] && I18N[lang][k]) || k;
function applyI18n() {
  document.documentElement.lang = lang; $('lang').value = lang;
  document.querySelectorAll('[data-i18n]').forEach(e => e.textContent = t(e.dataset.i18n));
  updatePresetHint();
}

// ---- API -------------------------------------------------------------------
async function api(method, url, body) {
  const res = await fetch(url, {
    method, headers: { 'X-WarpScout': '1', ...(body !== undefined ? { 'Content-Type': 'application/json' } : {}) },
    body: body !== undefined ? JSON.stringify(body) : undefined, credentials: 'same-origin'
  });
  if (res.status === 401) { showLogin(); throw new Error('unauthorized'); }
  const text = await res.text();
  let data = {}; try { data = text ? JSON.parse(text) : {}; } catch { data = { raw: text }; }
  if (!res.ok) { const e = new Error((data.error && data.error.message) || res.statusText); e.code = data.error && data.error.code; throw e; }
  return data;
}
function notice(msg, kind) {
  const n = $('notice'); n.textContent = msg; n.className = 'notice ' + (kind || ''); n.hidden = !msg;
  if (msg && kind !== 'error') setTimeout(() => { if (n.textContent === msg) n.hidden = true; }, 4000);
}
const fail = (e) => { if (e.message !== 'unauthorized') notice(e.message, 'error'); };

// ---- login -----------------------------------------------------------------
function showLogin() { $('login').hidden = false; $('login-token').focus(); }
$('login-form').addEventListener('submit', async (ev) => {
  ev.preventDefault(); $('login-error').hidden = true;
  try { await api('POST', '/api/login', { token: $('login-token').value }); $('login').hidden = true; $('login-token').value = ''; boot(); }
  catch (e) { $('login-error').textContent = e.message; $('login-error').hidden = false; }
});
$('logout').addEventListener('click', async () => { await fetch('/api/logout', { method: 'POST', headers: { 'X-WarpScout': '1' } }); showLogin(); });

// ---- tabs ------------------------------------------------------------------
document.querySelectorAll('#tabs button').forEach(b => b.addEventListener('click', () => selectTab(b.dataset.tab)));
function selectTab(name) {
  document.querySelectorAll('#tabs button').forEach(b => b.classList.toggle('active', b.dataset.tab === name));
  document.querySelectorAll('.tab').forEach(s => s.hidden = s.id !== 'tab-' + name);
  if (name === 'history') loadHistory();
  if (name === 'account') loadAccount();
  if (name === 'tools') loadTools();
}
$('lang').addEventListener('change', () => { lang = $('lang').value; try { localStorage.setItem('lang', lang); } catch {} applyI18n(); renderResults(); renderHistory(); loadAccount(); });

// ---- options ---------------------------------------------------------------
const list = (id) => $(id).value.split(',').map(s => s.trim()).filter(Boolean);
const num = (id) => parseInt($(id).value, 10) || 0;
function updatePresetHint() { $('preset-hint').textContent = t('hint_' + $('preset').value); }
$('preset').addEventListener('change', updatePresetHint);
$('telegramOnly').addEventListener('change', () => { if ($('telegramOnly').checked) $('telegram').checked = true; });
$('telegram').addEventListener('change', () => { if (!$('telegram').checked) $('telegramOnly').checked = false; });
$('expert').addEventListener('change', () => { $('expert-box').hidden = !$('expert').checked; });

// The form is sent as-is; the server resolves presets (core.ScanSpec.Resolve).
function buildSpec() {
  return {
    preset: $('preset').value, expert: $('expert').checked, protocol: $('protocol').value, ipv6: $('ipv6').checked,
    telegram: $('telegram').checked || $('telegramOnly').checked, telegramOnly: $('telegramOnly').checked,
    options: {
      protocol: $('protocol').value, innerProtocol: $('innerProtocol').value, ipv6: $('ipv6').checked,
      throughEndpoint: $('throughEndpoint').value.trim(), port: num('port'), timeoutSec: num('timeoutSec') || 2,
      jobs: num('jobs') || 10, tunnelPingCount: num('tunnelPingCount'), mtu: num('mtu'), dns: list('dns'),
      customTarget: $('customTarget').value.trim(), pingTarget: $('pingTarget').value.trim(),
      includeCountries: list('includeCountries'), includeNodes: list('includeNodes'),
      excludeCountries: list('excludeCountries'), excludeNodes: list('excludeNodes'),
      bestBy: $('bestBy').value, sweepPorts: $('sweepPorts').value,
      awgJunkCount: num('awgJunkCount'), awgJunkMin: num('awgJunkMin'), awgJunkMax: num('awgJunkMax'),
      awgAutoI1: $('awgAutoI1').checked, awgI1: $('awgI1').value.trim(), masqueSni: $('masqueSni').value.trim(),
      masqueAttempts: 3, speedTest: $('speedTest').checked
    }
  };
}
// Restore the form from a stored spec, or from legacy resolved options.
function fillForm(spec) {
  const o = spec.options || spec;
  $('preset').value = spec.preset || (o.full ? 'full' : o.tunnelPingCount >= 10 ? 'durable' : 'standard'); updatePresetHint();
  $('protocol').value = spec.protocol || o.protocol || 'awg'; $('ipv6').checked = !!(spec.ipv6 ?? o.ipv6);
  $('telegram').checked = !!(spec.telegram ?? o.telegram); $('telegramOnly').checked = !!(spec.telegramOnly ?? o.telegramOnly);
  $('expert').checked = spec.expert !== undefined ? !!spec.expert : true; $('expert-box').hidden = !$('expert').checked;
  const set = (id, v) => { $(id).value = v == null ? '' : v; };
  set('innerProtocol', o.innerProtocol || 'wg'); set('throughEndpoint', o.throughEndpoint); set('port', o.port || 0); set('timeoutSec', o.timeoutSec || 2);
  set('jobs', o.jobs || 10); set('tunnelPingCount', o.tunnelPingCount || 0); set('mtu', o.mtu || 0); set('dns', (o.dns || []).join(', '));
  set('customTarget', o.customTarget); set('pingTarget', o.pingTarget);
  set('includeCountries', (o.includeCountries || []).join(', ')); set('includeNodes', (o.includeNodes || []).join(', '));
  set('excludeCountries', (o.excludeCountries || []).join(', ')); set('excludeNodes', (o.excludeNodes || []).join(', '));
  set('bestBy', o.bestBy || 'ping'); set('sweepPorts', o.sweepPorts || '');
  set('awgJunkCount', o.awgJunkCount || 0); set('awgJunkMin', o.awgJunkMin || 0); set('awgJunkMax', o.awgJunkMax || 0);
  $('awgAutoI1').checked = o.awgAutoI1 !== false && !o.awgI1; set('awgI1', o.awgI1); set('masqueSni', o.masqueSni); $('speedTest').checked = !!o.speedTest;
}

// ---- running operations ----------------------------------------------------
let es = null, currentOp = '', rows = [], lastReport = null, seenSeq = 0;
const OP_TITLES = { register:'register', scan:'scan_start', 'find-junk':'tool_junk', 'find-sni':'tool_sni', socks:'SOCKS5' };

async function run(operation, body) {
  try {
    await api('POST', '/api/run', { operation, ...body });
    startOpUi(operation); seenSeq = 0; subscribe(0);
  } catch (e) { if (e.code === 'account_required') { notice(t('need_account'), 'error'); selectTab('account'); } else fail(e); }
}
function startOpUi(op) {
  currentOp = op; $('op-panel').hidden = false;
  $('op-title').textContent = OP_TITLES[op] && I18N[lang][OP_TITLES[op]] ? t(OP_TITLES[op]) : (OP_TITLES[op] || op);
  $('op-phase').textContent = t('running'); $('op-progress').removeAttribute('value'); $('op-message').textContent = '';
  setBusy(true);
  if (op === 'scan') { rows = []; lastReport = null; renderResults(); $('results-card').hidden = false; $('report-btn').hidden = true; }
}
function setBusy(b) { ['scan-start','junk-start','sni-start','socks-start','register'].forEach(id => $(id).disabled = b); }
function subscribe(after) {
  if (es) es.close();
  es = new EventSource('/api/events?after=' + after);
  es.addEventListener('job', (m) => { try { const j = JSON.parse(m.data); if (j.operation !== currentOp) startOpUi(j.operation); } catch {} });
  es.onmessage = (m) => { seenSeq = parseInt(m.lastEventId, 10) || seenSeq; try { onEvent(JSON.parse(m.data)); } catch (e) { console.error(e); } };
  es.addEventListener('end', () => { es.close(); es = null; finishOpUi(); });
  es.onerror = () => { /* EventSource retries on its own; auth loss is detected by the next api() call */ };
}
function onEvent(ev) {
  if (ev.type === 'progress' || ev.type === 'endpoint') {
    const bar = $('op-progress');
    if (ev.total > 0) { bar.max = ev.total; bar.value = ev.completed; } else bar.removeAttribute('value');
    $('op-phase').textContent = ev.phase ? ev.phase + (ev.total > 0 ? ` ${ev.completed}/${ev.total}` : '') : t('running');
    if (ev.message) $('op-message').textContent = ev.message;
    if (ev.endpoint && ev.operation === 'scan') { rows.push(ev.endpoint); renderResults(true); }
    if (ev.operation === 'socks' && ev.phase === 'listening') $('socks-status').textContent = ev.message;
  } else if (ev.type === 'completed') onCompleted(ev);
  else if (ev.type === 'error') onError(ev.error || {});
}
function onCompleted(ev) {
  const p = ev.payload;
  $('op-phase').textContent = t('done');
  if (ev.operation === 'scan') { lastReport = p; rows = p.results || []; renderResults(); }
  else if (ev.operation === 'register') { notice(t('acc_yes'), 'ok'); loadAccount(); }
  else if (ev.operation === 'find-junk' || ev.operation === 'find-sni') loadTools();
  else if (ev.operation === 'socks') $('socks-status').textContent = '';
}
function onError(err) {
  const code = err.code || '';
  $('op-phase').textContent = code === 'canceled' ? t('canceled') : (err.message || code);
  if (code !== 'canceled') notice((code ? code + ': ' : '') + (err.message || ''), 'error');
  if (currentOp === 'socks') $('socks-status').textContent = '';
}
function finishOpUi() { setBusy(false); const bar = $('op-progress'); bar.max = 100; bar.value = 100; $('op-cancel').disabled = false; setTimeout(() => { if (!es) $('op-panel').hidden = true; }, 4000); }
$('op-cancel').addEventListener('click', () => api('POST', '/api/cancel', {}).catch(fail));

let lastSpec = null;
$('scan-start').addEventListener('click', () => { lastSpec = buildSpec(); run('scan', { spec: lastSpec }); });
$('junk-start').addEventListener('click', () => { const spec = buildSpec(); spec.protocol = 'awg'; run('find-junk', { spec }); });
$('sni-start').addEventListener('click', () => { const spec = buildSpec(); spec.protocol = $('sni-protocol').value; run('find-sni', { spec }); });
$('socks-start').addEventListener('click', () => {
  const endpoint = $('socks-endpoint').value.trim(); if (!endpoint) { notice('Endpoint?', 'error'); return; }
  run('socks', { spec: buildSpec(), socks: { endpoint, port: num('socks-port') || 1080, protocol: $('socks-protocol').value } });
});
$('register').addEventListener('click', () => run('register', { payload: { fresh: $('reg-fresh').checked, proxy: $('reg-proxy').value.trim(), timeoutSec: num('reg-timeout') || 5, ipv6: false } }));

// ---- results table ---------------------------------------------------------
const flagOn = () => { try { return localStorage.getItem('flags') !== '0'; } catch { return true; } };
const flagOf = (cc) => /^[A-Za-z]{2}$/.test(cc || '') ? String.fromCodePoint(...[...cc.toUpperCase()].map(c => 0x1F1E6 + c.charCodeAt(0) - 65)) + ' ' : '';
$('flags').checked = flagOn();
$('flags').addEventListener('change', () => { try { localStorage.setItem('flags', $('flags').checked ? '1' : '0'); } catch {} renderResults(); });
function tgCell(r) {
  if (!r.telegram) return ['', '—'];
  if (r.telegram.ok) return ['ok', `✓ ${r.telegram.rttMs.toFixed(0)} ms`];
  return ['bad', r.telegram.dcsReached > 0 ? `${r.telegram.dcsReached}/${r.telegram.dcsTotal}` : t('tg_blocked')];
}
const f1 = (v, s) => v ? v.toFixed(1) + (s || '') : '—';
function sortedRows() {
  // Telegram reachability outranks latency (as in the core); under "only with
  // Telegram" every row reached it, so Telegram response time ranks instead.
  const tgOnly = $('telegramOnly').checked;
  const tgOk = (r) => !!(r.telegram && r.telegram.ok);
  const score = (r) => (r.working ? 0 : 1e7) + (tgOk(r) ? 0 : 1e6) + (tgOnly && tgOk(r) ? r.telegram.rttMs : (r.tunnelPingMs || r.endpointPingMs || 9999));
  return [...rows].sort((a, b) => score(a) - score(b));
}
function renderResults(live) {
  const body = $('results').tBodies[0]; body.textContent = '';
  const tgChecked = rows.some(r => r.telegram);
  const sorted = sortedRows();
  sorted.forEach((r, i) => {
    const tr = document.createElement('tr'); if (i === 0 && r.working) tr.className = 'best';
    const status = r.working ? (r.durable ? ['ok', t('working')] : ['warn', t('torn')]) : ['bad', t('failed')];
    const place = (flagOn() ? flagOf(r.country) : '') + ([r.country, r.node, r.nodeLocation || r.region].filter(Boolean).join(' · ') || '—');
    const cells = [[i + 1, ''], [r.endpoint, ''], [status[1], status[0]], [place, ''], [f1(r.endpointPingMs), 'num'], [f1(r.tunnelPingMs), 'num'],
      [r.lossPercent ? r.lossPercent.toFixed(0) + '%' : '—', 'num'], [f1(r.speedMbps), 'num'], tgCell(r).reverse()];
    cells.forEach(([v, c], idx) => { const td = document.createElement('td'); td.textContent = v; if (c) td.className = c; if (idx === 8) td.hidden = !tgChecked; tr.appendChild(td); });
    const td = document.createElement('td');
    if (r.working && !live) {
      const b = document.createElement('button'); b.className = 'ghost'; b.textContent = t('config');
      b.addEventListener('click', () => openConfig(r)); td.appendChild(b);
    }
    tr.appendChild(td); body.appendChild(tr);
  });
  const working = rows.filter(r => r.working).length;
  const checked = rows.some(r => r.telegram), tgWorking = rows.filter(r => r.telegram && r.telegram.ok).length;
  document.querySelectorAll('.tg-col').forEach(e => { e.hidden = !checked; });
  $('tg-hint').hidden = !checked;
  $('results-meta').textContent = rows.length ? `${t('found')}: ${working} ${t('of')} ${rows.length}` + (checked ? ` · Telegram: ${tgWorking}` : '') : '';
  const done = !live && working > 0;
  $('report-btn').hidden = !(lastReport && !live); $('best-btn').hidden = !done; $('csv-btn').hidden = !done;
}
function bestRow() { const r = sortedRows()[0]; return r && r.working ? r : null; }
$('best-btn').addEventListener('click', async () => {
  const r = bestRow(); if (!r) return;
  try { await navigator.clipboard.writeText(r.endpoint); notice(t('copied') + ': ' + r.endpoint, 'ok'); } catch { showModal('best.txt', r.endpoint, null); }
});
const csvCell = (v) => { v = String(v ?? ''); return /[",\n]/.test(v) ? '"' + v.replace(/"/g, '""') + '"' : v; };
$('csv-btn').addEventListener('click', () => {
  const head = ['endpoint','working','durable','country','node','region','nodeLocation','exitIp','endpointPingMs','tunnelPingMs','lossPercent','speedMbps'];
  const tgHead = ['telegramOk','telegramDcs','telegramMs'];
  const tgVals = (r) => r.telegram ? [r.telegram.ok, r.telegram.dcsReached, r.telegram.ok ? r.telegram.rttMs : ''] : ['', '', ''];
  const lines = [head.concat(tgHead).join(',')].concat(sortedRows().map(r => head.map(k => csvCell(r[k])).concat(tgVals(r).map(csvCell)).join(',')));
  const a = document.createElement('a'); a.href = URL.createObjectURL(new Blob([lines.join('\n') + '\n'], { type: 'text/csv' }));
  a.download = 'warpscout-results.csv'; a.click(); setTimeout(() => URL.revokeObjectURL(a.href), 1000);
});
$('report-btn').addEventListener('click', async () => {
  try { const r = await api('POST', '/api/render', { operation: 'render-report', payload: { report: lastReport } }); showModal('report.txt', r.result, null); } catch (e) { fail(e); }
});

// ---- modal -----------------------------------------------------------------
let modalEndpoint = null, modalName = 'file.txt';
function showModal(title, text, endpoint) {
  $('modal').hidden = false; $('modal-title').textContent = title; $('modal-body').textContent = text; modalEndpoint = endpoint;
  $('modal-format').hidden = !endpoint; modalName = title.replace(/[^\w.\-]+/g, '_') || 'file.txt';
}
async function openConfig(endpoint) {
  modalEndpoint = endpoint; $('modal-format').hidden = false;
  const proto = (lastReport && lastReport.protocol) || $('protocol').value;
  $('modal-format').value = proto === 'wg' ? 'wireguard' : proto.startsWith('masque') ? 'usque' : 'amneziawg';
  await renderModalConfig();
}
async function renderModalConfig() {
  const format = $('modal-format').value, spec = lastSpec || buildSpec();
  try {
    const r = await api('POST', '/api/render', { operation: 'render-config', spec, payload: { endpoint: modalEndpoint, format } });
    const ext = { wireguard: 'conf', amneziawg: 'conf', usque: 'json', mihomo: 'yaml' }[format];
    showModal(`${modalEndpoint.endpoint.replace(/[:\[\]]/g, '_')}.${ext}`, r.result, modalEndpoint);
  } catch (e) { fail(e); }
}
$('modal-format').addEventListener('change', renderModalConfig);
$('modal-close').addEventListener('click', () => { $('modal').hidden = true; });
$('modal').addEventListener('click', (e) => { if (e.target === $('modal')) $('modal').hidden = true; });
$('modal-copy').addEventListener('click', async () => { try { await navigator.clipboard.writeText($('modal-body').textContent); notice(t('copied'), 'ok'); } catch { const r = document.createRange(); r.selectNodeContents($('modal-body')); getSelection().removeAllRanges(); getSelection().addRange(r); } });
$('modal-download').addEventListener('click', () => {
  const a = document.createElement('a'); a.href = URL.createObjectURL(new Blob([$('modal-body').textContent], { type: 'text/plain' }));
  a.download = modalName; a.click(); setTimeout(() => URL.revokeObjectURL(a.href), 1000);
});

// ---- saved tool results ----------------------------------------------------
let tools = {};
async function loadTools() {
  try { tools = await api('GET', '/api/tools'); } catch (e) { return fail(e); }
  const j = tools['find-junk'], n = tools['find-sni'];
  $('junk-saved').hidden = !j; $('sni-saved').hidden = !n;
  const when = (x) => `${t('saved_at')}: ${new Date(x.at).toLocaleString(lang)}`;
  if (j) { $('junk-saved-meta').textContent = when(j); $('junk-result').textContent = `Jc=${j.payload.junkCount}  Jmin=${j.payload.junkMin}  Jmax=${j.payload.junkMax}\nI1: ${j.payload.i1 || '—'}`; }
  if (n) { $('sni-saved-meta').textContent = when(n); $('sni-result').textContent = `SNI: ${n.payload.sni || '—'}  (${n.payload.protocol || ''})`; }
}
$('junk-apply').addEventListener('click', () => {
  const j = tools['find-junk']; if (!j) return; const p = j.payload;
  $('protocol').value = 'awg'; $('expert').checked = true; $('expert-box').hidden = false;
  $('awgJunkCount').value = p.junkCount || 0; $('awgJunkMin').value = p.junkMin || 0; $('awgJunkMax').value = p.junkMax || 0;
  $('awgAutoI1').checked = !p.i1; $('awgI1').value = p.i1 || '';
  notice(t('applied'), 'ok'); selectTab('scan');
});
$('sni-apply').addEventListener('click', () => {
  const n = tools['find-sni']; if (!n || !n.payload.sni) return;
  $('protocol').value = n.payload.protocol || 'masque'; $('expert').checked = true; $('expert-box').hidden = false;
  $('masqueSni').value = n.payload.sni; notice(t('applied'), 'ok'); selectTab('scan');
});

// ---- history ---------------------------------------------------------------
let historyItems = [];
async function loadHistory() { try { historyItems = (await api('GET', '/api/history')).items || []; renderHistory(); } catch (e) { fail(e); } }
function renderHistory() {
  const body = $('history').tBodies[0]; body.textContent = ''; $('history-empty').hidden = historyItems.length > 0;
  historyItems.forEach(h => {
    const tr = document.createElement('tr');
    const when = new Date(h.createdAt).toLocaleString(lang);
    [when, (h.preset || '') + ' · ' + h.protocol, `${h.working}/${h.total}` + (h.telegramOk ? ` · TG ${h.telegramOk}` : ''), h.best ? `${t('best')}: ${h.best}${h.bestPingMs ? ' (' + h.bestPingMs.toFixed(0) + ' ms)' : ''}` : '—']
      .forEach(v => { const td = document.createElement('td'); td.textContent = v; tr.appendChild(td); });
    const td = document.createElement('td');
    const open = document.createElement('button'); open.className = 'ghost'; open.textContent = t('open'); open.addEventListener('click', () => openHistory(h.id));
    const del = document.createElement('button'); del.className = 'ghost danger-text'; del.textContent = '✕';
    del.addEventListener('click', async () => { try { await api('DELETE', '/api/history/' + h.id); loadHistory(); } catch (e) { fail(e); } });
    td.append(open, ' ', del); tr.appendChild(td); body.appendChild(tr);
  });
}
async function openHistory(id) {
  try {
    const entry = await api('GET', '/api/history/' + id);
    lastReport = entry.report; rows = lastReport.results || [];
    if (entry.spec) { lastSpec = entry.spec; fillForm(entry.spec); }
    else if (entry.options) { fillForm({ ...entry.options, preset: entry.preset }); lastSpec = buildSpec(); }
    selectTab('scan'); $('results-card').hidden = false; renderResults();
  } catch (e) { fail(e); }
}
$('history-clear').addEventListener('click', async () => { if (confirm(t('confirm_clear'))) { try { await api('DELETE', '/api/history'); loadHistory(); } catch (e) { fail(e); } } });

// ---- account ---------------------------------------------------------------
async function loadAccount() {
  try { const a = await api('GET', '/api/account'); $('account-status').textContent = a.registered ? t('acc_yes') : t('acc_no');
    $('account-export').disabled = $('account-delete').disabled = !a.registered; } catch (e) { fail(e); }
}
$('account-export').addEventListener('click', async () => {
  try {
    const st = await api('GET', '/api/state');
    const token = st.authRequired ? prompt(t('confirm_token')) : '';
    if (st.authRequired && !token) return;
    const res = await fetch('/api/account/export', { method: 'POST', headers: { 'X-WarpScout': '1', 'Content-Type': 'application/json' }, body: JSON.stringify({ token }), credentials: 'same-origin' });
    if (!res.ok) { const d = await res.json().catch(() => ({})); throw new Error((d.error && d.error.message) || res.statusText); }
    const a = document.createElement('a'); a.href = URL.createObjectURL(await res.blob()); a.download = 'warpscout-account.json'; a.click();
    setTimeout(() => URL.revokeObjectURL(a.href), 1000);
  } catch (e) { fail(e); }
});
$('account-delete').addEventListener('click', async () => { if (confirm(t('confirm_delete_acc'))) { try { await api('DELETE', '/api/account'); loadAccount(); } catch (e) { fail(e); } } });
$('account-import-btn').addEventListener('click', async () => {
  try { await api('PUT', '/api/account', { rawJson: $('account-import').value.trim() }); $('account-import').value = ''; notice(t('acc_yes'), 'ok'); loadAccount(); } catch (e) { fail(e); }
});

// ---- boot ------------------------------------------------------------------
async function boot() {
  try {
    const s = await api('GET', '/api/state');
    $('logout').hidden = false;
    if (s.job) {
      startOpUi(s.job.operation);
      if (!s.job.running) { setBusy(false); }
      subscribe(0);
    }
    loadTools();
    if (!s.hasAccount) { selectTab('account'); loadAccount(); }
  } catch (e) { if (e.message !== 'unauthorized') notice(e.message, 'error'); }
}
applyI18n(); boot();
