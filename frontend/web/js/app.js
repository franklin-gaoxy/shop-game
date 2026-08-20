// app.js Web 前端主逻辑
"use strict";

// ==================== 全局状态 ====================
const state = {
  user: null,
  shopPage: 1,
  shopData: null,
  whPage: 1,
  whData: null,
  txPage: 1,
  currentPage: "home",
};

const PAGE_SIZE = 20; // 商城/仓库/交易记录默认每页 20 条

// ==================== 工具函数 ====================
const $ = (id) => document.getElementById(id);

function esc(s) {
  return String(s ?? "").replace(/[&<>"']/g, (c) => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
  }[c]));
}

function money(v) {
  return Number(v || 0).toLocaleString("zh-CN", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

let toastTimer = null;
function toast(msg, type) {
  const el = $("toast");
  el.textContent = msg;
  el.className = "toast " + (type || "");
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => el.classList.add("hidden"), 2600);
}

function expireDaysText(days) {
  if (days === -1) return "永久";
  if (days === 0) return "不适用";
  return days + " 天";
}

function expireDayText(day) {
  return day === -1 ? "永久" : "第 " + day + " 天";
}

// ==================== 视图切换 ====================
function showView(name) {
  $("auth-view").classList.toggle("hidden", name !== "auth");
  $("main-view").classList.toggle("hidden", name !== "main");
}

function updateTopbar() {
  if (!state.user) return;
  $("topbar-user").textContent =
    state.user.username + " ｜ 第 " + state.user.day + " 天 ｜ 余额 " + money(state.user.money);
}

async function refreshUser() {
  try {
    state.user = await API.me();
    updateTopbar();
  } catch (e) {
    /* 忽略 */
  }
}

function switchPage(name) {
  state.currentPage = name;
  document.querySelectorAll(".nav-link").forEach((b) =>
    b.classList.toggle("active", b.dataset.page === name)
  );
  document.querySelectorAll(".page").forEach((p) =>
    p.classList.toggle("hidden", p.id !== "page-" + name)
  );
  switch (name) {
    case "home": loadHome(); break;
    case "shop": loadShop(); break;
    case "special": loadSpecial(); break;
    case "warehouse": loadWarehouse(); break;
    case "transactions": loadTransactions(); break;
  }
}

// ==================== 登录 / 注册 ====================
function initAuth() {
  // tab 切换：默认展示登录表单
  $("tab-login").addEventListener("click", () => switchAuthTab("login"));
  $("tab-register").addEventListener("click", () => switchAuthTab("register"));

  $("login-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const f = e.target;
    const msg = $("login-msg");
    msg.textContent = "";
    try {
      state.user = await API.login(f.username.value.trim(), f.password.value);
      enterMain();
    } catch (err) {
      msg.textContent = err.message;
    }
  });

  $("register-form").addEventListener("submit", async (e) => {
    e.preventDefault();
    const f = e.target;
    const msg = $("register-msg");
    msg.textContent = "";
    try {
      state.user = await API.register(
        f.username.value.trim(), f.password.value, f.key.value.trim()
      );
      enterMain();
    } catch (err) {
      msg.textContent = err.message;
    }
  });

  $("auth-server-edit").addEventListener("click", () => openSettings());
}

function switchAuthTab(which) {
  $("tab-login").classList.toggle("active", which === "login");
  $("tab-register").classList.toggle("active", which === "register");
  $("login-form").classList.toggle("hidden", which !== "login");
  $("register-form").classList.toggle("hidden", which !== "register");
}

function enterMain() {
  showView("main");
  updateTopbar();
  switchPage("home");
  toast("欢迎回来，" + state.user.username, "success");
}

// ==================== 设置弹窗（全局服务器地址） ====================
function openSettings() {
  $("settings-server").value = API.base;
  $("settings-msg").textContent = "";
  $("settings-modal").classList.remove("hidden");
}

function initSettings() {
  $("btn-settings").addEventListener("click", openSettings);
  $("settings-close").addEventListener("click", () =>
    $("settings-modal").classList.add("hidden")
  );
  $("settings-save").addEventListener("click", async () => {
    const v = $("settings-server").value.trim();
    if (!/^https?:\/\/.+/.test(v)) {
      $("settings-msg").textContent = "地址需以 http:// 或 https:// 开头";
      return;
    }
    API.base = v;
    $("settings-msg").className = "form-msg ok";
    $("settings-msg").textContent = "已保存，正在测试连接…";
    try {
      await API.health();
      $("settings-msg").textContent = "已保存，连接成功";
      $("auth-server-addr").textContent = API.base;
      setTimeout(() => $("settings-modal").classList.add("hidden"), 600);
    } catch (e) {
      $("settings-msg").className = "form-msg";
      $("settings-msg").textContent = "已保存，但连接失败：" + e.message;
    }
  });
}

// ==================== 首页 ====================
function renderStorage(prefix, data) {
  const pct = Math.min(100, Number(data.usage_percent) || 0);
  const set = (id, v) => {
    const el = $(id);
    if (el) el.textContent = v;
  };
  set(prefix + "-text", pct.toFixed(1) + "%");
  const bar = $(prefix + "-bar");
  if (bar) bar.style.width = pct + "%";
  set(prefix + "-used", data.used);
  set(prefix + "-free", data.free);
  set(prefix + "-cap", data.capacity);
}

async function loadHome() {
  try {
    const d = await API.home();
    state.user = state.user || {};
    state.user.money = d.money;
    state.user.day = d.day;
    $("home-money").textContent = money(d.money);
    $("home-day").textContent = "第 " + d.day + " 天";
    renderStorage("home-normal", d.normal);
    renderStorage("home-cold", d.cold);
    updateTopbar();
  } catch (e) {
    toast(e.message, "error");
  }
}

async function doTomorrow() {
  const btn = $("btn-tomorrow");
  btn.disabled = true;
  try {
    const res = await API.tomorrow();
    state.user.day = res.day;
    toast("已更新到明天，当前第 " + res.day + " 天", "success");
    const events = [];
    (res.triggered_crits || []).forEach((c) =>
      events.push(
        '<div class="event-line"><span class="tag crit">暴击</span>' +
          esc(c.description) + "（影响商品：" + esc((c.products || []).join("、")) + "）</div>"
      )
    );
    (res.expired_items || []).forEach((e2) =>
      events.push(
        '<div class="event-line"><span class="tag expire">过期</span>' +
          esc(e2.product_name) + " ×" + e2.quantity + " 已过期销毁（损失成本 " + money(e2.value) + "）</div>"
      )
    );
    if (res.expired_purchases > 0) {
      events.push(
        '<div class="event-line"><span class="tag space">仓库</span>有 ' +
          res.expired_purchases + " 笔已购买的仓库空间到期失效</div>"
      );
    }
    $("tomorrow-events").innerHTML =
      events.length ? events.join("") : '<div class="event-line">今日无暴击与过期事件</div>';
    $("tomorrow-result").style.display = "";
    await loadHome();
  } catch (e) {
    toast(e.message, "error");
  } finally {
    btn.disabled = false;
  }
}

// ==================== 商城 ====================
async function loadShop() {
  try {
    const d = await API.products();
    state.shopData = d;
    $("shop-day").textContent = "第 " + d.day + " 天价格";
    renderShop();
  } catch (e) {
    toast(e.message, "error");
  }
}

function renderShop() {
  const d = state.shopData;
  const total = (d.products || []).length;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  if (state.shopPage > totalPages) state.shopPage = totalPages;
  const rows = d.products.slice((state.shopPage - 1) * PAGE_SIZE, state.shopPage * PAGE_SIZE);

  $("shop-tbody").innerHTML = rows.length
    ? rows
        .map(
          (p) => `<tr>
      <td class="freeze">${esc(p.name)}</td>
      <td>${esc(p.category)}</td>
      <td class="trend-up" style="color:var(--primary)">${money(p.price)}</td>
      <td>${p.crit_applied ? '<span class="trend-up">是</span>' : "-"}</td>
      <td>${money(p.min_price)} ~ ${money(p.max_price)}</td>
      <td>${p.size}</td>
      <td>${esc(storageTypeName(p.storage_type))}</td>
      <td>${expireDaysText(p.normal_expire_days)}</td>
      <td>${expireDaysText(p.cold_expire_days)}</td>
      <td><button class="btn small primary" data-buy="${p.id}">购买</button></td>
    </tr>`
        )
        .join("")
    : '<tr><td class="freeze" colspan="10">（暂无商品）</td></tr>';

  $("shop-tbody")
    .querySelectorAll("[data-buy]")
    .forEach((b) =>
      b.addEventListener("click", () => {
        const p = d.products.find((x) => x.id === Number(b.dataset.buy));
        if (p) openBuyModal(p);
      })
    );

  renderPager("shop-pager", state.shopPage, totalPages, (p) => {
    state.shopPage = p;
    renderShop();
  });
}

// ==================== 仓库 ====================
async function loadWarehouse() {
  try {
    const d = await API.warehouse();
    state.whData = d;
    $("wh-day").textContent = "第 " + d.day + " 天库存";
    renderWarehouse();
  } catch (e) {
    toast(e.message, "error");
  }
}

function renderWarehouse() {
  const d = state.whData;
  const total = (d.items || []).length;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));
  if (state.whPage > totalPages) state.whPage = totalPages;
  const rows = d.items.slice((state.whPage - 1) * PAGE_SIZE, state.whPage * PAGE_SIZE);

  $("wh-tbody").innerHTML = rows.length
    ? rows
        .map((it) => {
          // 均价高于今日价 → 绿色向上箭头 + 预计获利；低于 → 红色向下箭头 + 亏损
          let trend, rowCls;
          if (it.profit > 0) {
            trend = '<span class="trend-up">▲</span> 预计获利 ' + money(it.profit);
            rowCls = "row-up";
          } else if (it.profit < 0) {
            trend = '<span class="trend-down">▼</span> 预计亏损 ' + money(-it.profit);
            rowCls = "row-down";
          } else {
            trend = '<span class="trend-flat">—</span> 持平 0.00';
            rowCls = "";
          }
          return `<tr class="${rowCls}">
      <td class="freeze">${esc(it.product_name)}</td>
      <td>${esc(storageTypeName(it.storage_type))}</td>
      <td>${it.quantity}</td>
      <td>${money(it.avg_price)}</td>
      <td>${money(it.total_cost)}</td>
      <td>${money(it.today_price)}</td>
      <td>${trend}</td>
      <td>${expireDayText(it.expire_day)}</td>
      <td><button class="btn small primary" data-sell="${it.product_id}" data-storage="${it.storage_type}">卖出</button></td>
    </tr>`;
        })
        .join("")
    : '<tr><td class="freeze" colspan="9">（仓库为空）</td></tr>';

  $("wh-tbody")
    .querySelectorAll("[data-sell]")
    .forEach((b) =>
      b.addEventListener("click", () => {
        const item = d.items.find(
          (x) => x.product_id === Number(b.dataset.sell) && x.storage_type === Number(b.dataset.storage)
        );
        if (item) openSellModal(item);
      })
    );

  renderPager("wh-pager", state.whPage, totalPages, (p) => {
    state.whPage = p;
    renderWarehouse();
  });
}

// ==================== 特殊商城 ====================
async function loadSpecial() {
  try {
    const d = await API.warehouseSpace();
    $("sp-normal-unit").textContent = money(d.unit_price_normal);
    $("sp-cold-unit").textContent = money(d.unit_price_cold);
    renderStorage("sp-normal", {
      capacity: d.normal.capacity, used: d.normal.used, free: d.normal.free,
      usage_percent: d.normal.capacity ? (d.normal.used / d.normal.capacity) * 100 : 0,
    });
    renderStorage("sp-cold", {
      capacity: d.cold.capacity, used: d.cold.used, free: d.cold.free,
      usage_percent: d.cold.capacity ? (d.cold.used / d.cold.capacity) * 100 : 0,
    });

    $("sp-purchase-tbody").innerHTML = (d.purchases || []).length
      ? d.purchases
          .map(
            (p) => `<tr>
        <td class="freeze">${esc(storageTypeName(p.storage_type))}</td>
        <td>${p.size}</td>
        <td>${p.months}</td>
        <td>${money(p.unit_price)}</td>
        <td>第 ${p.start_day} 天</td>
        <td>第 ${p.expire_day} 天</td>
        <td>${p.expire_day > d.day ? '<span class="trend-up">生效中</span>' : '<span class="trend-down">已到期</span>'}</td>
      </tr>`
          )
          .join("")
      : '<tr><td class="freeze" colspan="7">（暂无购买记录）</td></tr>';
  } catch (e) {
    toast(e.message, "error");
  }
}

// ==================== 交易记录 ====================
async function loadTransactions() {
  try {
    const d = await API.transactions(state.txPage, PAGE_SIZE);
    const totalPages = Math.max(1, Math.ceil(d.total / PAGE_SIZE));
    $("tx-total").textContent = "共 " + d.total + " 条";

    const typeName = { buy: "买入", sell: "卖出", expire: "过期销毁", warehouse: "购买仓库" };
    const dirName = (x) => (Number(x) === 1 ? "支出" : "收入");
    const dirCls = (x) => (Number(x) === 1 ? "trend-down" : "trend-up");
    const fmtTime = (s) => (s || "").replace("T", " ").slice(0, 19);

    $("tx-tbody").innerHTML = d.list.length
      ? d.list
          .map(
            (t) => `<tr>
        <td class="freeze">${esc(t.product_name || "-")}</td>
        <td>第 ${t.day} 天</td>
        <td>${esc(typeName[t.type] || t.type)}</td>
        <td class="${dirCls(t.direction)}">${dirName(t.direction)}</td>
        <td>${t.quantity}</td>
        <td>${money(t.unit_price)}</td>
        <td class="${dirCls(t.direction)}">${Number(t.direction) === 1 ? "-" : "+"}${money(t.amount)}</td>
        <td>${money(t.balance_after)}</td>
        <td>${esc(fmtTime(t.created_at))}</td>
      </tr>`
          )
        .join("")
    : '<tr><td class="freeze" colspan="9">（暂无交易记录）</td></tr>';

    renderPager("tx-pager", state.txPage, totalPages, (p) => {
      state.txPage = p;
      loadTransactions();
    });
  } catch (e) {
    toast(e.message, "error");
  }
}

// ==================== 分页组件 ====================
function renderPager(containerId, page, totalPages, go) {
  const el = $(containerId);
  el.innerHTML = "";
  if (totalPages <= 1) return;

  const prev = document.createElement("button");
  prev.textContent = "上一页";
  prev.disabled = page <= 1;
  prev.addEventListener("click", () => go(page - 1));

  const info = document.createElement("span");
  info.innerHTML = '第 <span class="current">' + page + "</span> / " + totalPages + " 页";

  const next = document.createElement("button");
  next.textContent = "下一页";
  next.disabled = page >= totalPages;
  next.addEventListener("click", () => go(page + 1));

  el.append(prev, info, next);
}

// ==================== 买入 / 卖出弹窗 ====================
let tradeCtx = null; // {mode:'buy'|'sell', product}

function estimateText() {
  if (!tradeCtx) return "";
  const qty = Number($("trade-quantity").value) || 0;
  if (qty <= 0) return "";
  const price =
    tradeCtx.mode === "buy" ? tradeCtx.product.price : tradeCtx.product.today_price;
  const verb = tradeCtx.mode === "buy" ? "预计支出" : "预计收入";
  return verb + " " + money(price * qty) + "（单价 " + money(price) + " × " + qty + "）";
}

function openBuyModal(p) {
  tradeCtx = { mode: "buy", product: p };
  $("trade-title").textContent = "购买 " + p.name;
  const expire =
    p.storage_type === STORAGE_COLD ? p.cold_expire_days : p.normal_expire_days;
  $("trade-info").innerHTML =
    "种类：" + esc(p.category) +
    "｜今日价格：" + money(p.price) +
    "｜占用空间：" + p.size + "/个" +
    "<br>普通过期：" + expireDaysText(p.normal_expire_days) +
    "｜冷藏过期：" + expireDaysText(p.cold_expire_days);

  // 两种仓库均可存储时让用户选择
  const showStorage = p.storage_type === STORAGE_BOTH;
  $("trade-storage-label").classList.toggle("hidden", !showStorage);
  $("trade-storage").value = String(STORAGE_NORMAL);
  if (p.storage_type === STORAGE_COLD) $("trade-storage").value = String(STORAGE_COLD);

  $("trade-quantity").value = 1;
  $("trade-quantity").max = "";
  $("trade-estimate").textContent = estimateText();
  $("trade-msg").textContent = "";
  $("trade-modal").classList.remove("hidden");
}

function openSellModal(item) {
  tradeCtx = { mode: "sell", product: item };
  $("trade-title").textContent = "卖出 " + item.product_name;
  $("trade-info").innerHTML =
    "存储仓库：" + esc(storageTypeName(item.storage_type)) +
    "｜持有数量：" + item.quantity +
    "｜均价：" + money(item.avg_price) +
    "｜今日价格：" + money(item.today_price);
  $("trade-storage-label").classList.add("hidden");
  $("trade-quantity").value = item.quantity;
  $("trade-quantity").max = item.quantity;
  $("trade-estimate").textContent = estimateText();
  $("trade-msg").textContent = "";
  $("trade-modal").classList.remove("hidden");
}

async function confirmTrade() {
  const qty = Number($("trade-quantity").value);
  if (!Number.isInteger(qty) || qty <= 0) {
    $("trade-msg").textContent = "数量必须是正整数";
    return;
  }
  const btn = $("trade-confirm");
  btn.disabled = true;
  try {
    let res;
    if (tradeCtx.mode === "buy") {
      const storage = $("trade-storage-label").classList.contains("hidden")
        ? tradeCtx.product.storage_type === STORAGE_COLD ? STORAGE_COLD : STORAGE_NORMAL
        : Number($("trade-storage").value);
      res = await API.buy(tradeCtx.product.id, qty, storage);
      toast("购买成功：" + res.product_name + " ×" + res.quantity + "，支出 " + money(res.amount) + "，余额 " + money(res.balance), "success");
    } else {
      res = await API.sell(tradeCtx.product.product_id, qty, tradeCtx.product.storage_type);
      toast("卖出成功：" + res.product_name + " ×" + res.quantity + "，收入 " + money(res.amount) + "，余额 " + money(res.balance), "success");
    }
    $("trade-modal").classList.add("hidden");
    state.user.money = res.balance;
    updateTopbar();
    // 刷新当前页数据
    if (state.currentPage === "shop") loadShop();
    if (state.currentPage === "warehouse") loadWarehouse();
    if (state.currentPage === "home") loadHome();
  } catch (e) {
    $("trade-msg").textContent = e.message;
  } finally {
    btn.disabled = false;
  }
}

// ==================== 购买仓库空间弹窗 ====================
let spaceCtx = null; // {storageType, unitPrice, day}

function openSpaceModal(storageType, unitPrice, day) {
  spaceCtx = { storageType, unitPrice, day };
  $("space-title").textContent = "购买" + storageTypeName(storageType) + "空间";
  $("space-info").innerHTML =
    "单价：" + money(unitPrice) + " 元/单位/月（每月按 30 天）<br>最短 1 个月，最长 12 个月";
  $("space-size").value = 100;
  $("space-months").value = 1;
  updateSpaceEstimate();
  $("space-msg").textContent = "";
  $("space-modal").classList.remove("hidden");
}

function updateSpaceEstimate() {
  if (!spaceCtx) return;
  const size = Number($("space-size").value) || 0;
  const months = Number($("space-months").value) || 0;
  if (size <= 0 || months <= 0) {
    $("space-estimate").textContent = "";
    return;
  }
  const cost = spaceCtx.unitPrice * size * months;
  $("space-estimate").textContent =
    "预计支出 " + money(cost) + "（" + size + " 空间 × " + months + " 个月），到期：第 " +
    (spaceCtx.day + months * 30) + " 天";
}

async function confirmSpace() {
  const size = Number($("space-size").value);
  const months = Number($("space-months").value);
  if (!Number.isInteger(size) || size <= 0) {
    $("space-msg").textContent = "空间大小必须是正整数";
    return;
  }
  if (!Number.isInteger(months) || months < 1 || months > 12) {
    $("space-msg").textContent = "月数必须是 1-12 的整数";
    return;
  }
  const btn = $("space-confirm");
  btn.disabled = true;
  try {
    await API.warehouseBuy(spaceCtx.storageType, size, months);
    toast("购买成功：" + storageTypeName(spaceCtx.storageType) + " " + size + " 空间 × " + months + " 个月", "success");
    $("space-modal").classList.add("hidden");
    loadSpecial();
    refreshUser();
  } catch (e) {
    $("space-msg").textContent = e.message;
  } finally {
    btn.disabled = false;
  }
}

// ==================== 退出登录 ====================
async function doLogout() {
  try {
    await API.logout();
  } catch (e) {
    /* 忽略 */
  }
  state.user = null;
  state.txPage = 1;
  showView("auth");
  toast("已退出登录", "success");
}

// ==================== 初始化 ====================
function init() {
  $("auth-server-addr").textContent = API.base;
  initAuth();
  initSettings();

  // 导航
  document.querySelectorAll(".nav-link").forEach((b) =>
    b.addEventListener("click", () => switchPage(b.dataset.page))
  );

  // 首页
  $("btn-tomorrow").addEventListener("click", doTomorrow);

  // 交易弹窗
  $("trade-quantity").addEventListener("input", () => {
    $("trade-estimate").textContent = estimateText();
  });
  $("trade-confirm").addEventListener("click", confirmTrade);
  $("trade-cancel").addEventListener("click", () => $("trade-modal").classList.add("hidden"));

  // 仓库空间弹窗
  document.querySelectorAll("[data-buy-space]").forEach((b) =>
    b.addEventListener("click", async () => {
      try {
        const d = await API.warehouseSpace();
        const storageType = Number(b.dataset.buySpace);
        const unitPrice = storageType === STORAGE_COLD ? d.unit_price_cold : d.unit_price_normal;
        openSpaceModal(storageType, unitPrice, d.day);
      } catch (e) {
        toast(e.message, "error");
      }
    })
  );
  $("space-size").addEventListener("input", updateSpaceEstimate);
  $("space-months").addEventListener("input", updateSpaceEstimate);
  $("space-confirm").addEventListener("click", confirmSpace);
  $("space-cancel").addEventListener("click", () => $("space-modal").classList.add("hidden"));

  // 退出
  $("btn-logout").addEventListener("click", doLogout);

  // 已有 token 时尝试恢复会话
  if (API.token) {
    API.me()
      .then((u) => {
        state.user = u;
        enterMain();
      })
      .catch(() => {
        API.token = "";
        showView("auth");
      });
  } else {
    showView("auth");
  }
}

document.addEventListener("DOMContentLoaded", init);
