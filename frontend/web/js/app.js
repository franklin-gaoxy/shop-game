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
  adminUserPage: 1,
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
  // 仅管理员显示"密钥和用户管理"菜单
  $("btn-admin").classList.toggle("hidden", !state.user.is_admin);
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
  $("btn-admin").classList.toggle("active", name === "admin");
  document.querySelectorAll(".page").forEach((p) =>
    p.classList.toggle("hidden", p.id !== "page-" + name)
  );
  switch (name) {
    case "home": loadHome(); break;
    case "shop": loadShop(); break;
    case "special": loadSpecial(); break;
    case "warehouse": loadWarehouse(); break;
    case "transactions": loadTransactions(); break;
    case "admin": loadAdmin(); break;
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
  // 首页按钮与顶栏按钮同时禁用，防止重复提交
  const btns = [$("btn-tomorrow"), $("btn-tomorrow-top")].filter(Boolean);
  btns.forEach((b) => (b.disabled = true));
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
    // 刷新当前所在页面的数据（价格/库存/空间随天数变化）
    switch (state.currentPage) {
      case "shop": await loadShop(); break;
      case "warehouse": await loadWarehouse(); break;
      case "special": await loadSpecial(); break;
      case "transactions": await loadTransactions(); break;
      default: await loadHome();
    }
  } catch (e) {
    toast(e.message, "error");
  } finally {
    btns.forEach((b) => (b.disabled = false));
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

// ==================== 密钥和用户管理（仅 admin） ====================
async function loadAdmin() {
  if (!state.user || !state.user.is_admin) {
    switchPage("home");
    return;
  }
  await Promise.all([loadAdminUsers(), loadAdminKeys()]);
}

// ---------- 用户管理 ----------
async function loadAdminUsers() {
  try {
    const d = await API.adminUsers(state.adminUserPage, PAGE_SIZE);
    $("admin-user-total").textContent = "共 " + d.total + " 人";
    const totalPages = Math.max(1, Math.ceil(d.total / PAGE_SIZE));
    if (state.adminUserPage > totalPages) state.adminUserPage = totalPages;

    const fmtTime = (s) => (s || "").replace("T", " ").slice(0, 19);
    $("admin-user-tbody").innerHTML = d.list.length
      ? d.list
          .map((u) => {
            const self = u.id === state.user.id;
            const statusHtml = u.status === 1
              ? '<span class="trend-up">正常</span>'
              : '<span class="trend-down">已禁用</span>';
            // 操作列：管理员自身不可禁用/删除
            const actions = self
              ? '<span class="trend-flat">当前账号</span>'
              : (u.status === 1
                  ? '<button class="btn small danger" data-ustatus="0" data-uid="' + u.id + '" data-uname="' + esc(u.username) + '">禁用</button> '
                  : '<button class="btn small" data-ustatus="1" data-uid="' + u.id + '" data-uname="' + esc(u.username) + '">启用</button> ') +
                '<button class="btn small danger" data-udel="' + u.id + '" data-uname="' + esc(u.username) + '">删除</button>';
            return `<tr>
        <td class="freeze">${esc(u.username)}</td>
        <td>${u.is_admin ? "管理员" : "普通用户"}</td>
        <td>${statusHtml}</td>
        <td>第 ${u.day} 天</td>
        <td>${money(u.money)}</td>
        <td>${u.key_id || "-"}</td>
        <td>${esc(fmtTime(u.created_at))}</td>
        <td>${actions}</td>
      </tr>`;
          })
          .join("")
      : '<tr><td class="freeze" colspan="8">（暂无用户）</td></tr>';

    // 禁用/启用/删除事件
    $("admin-user-tbody").querySelectorAll("[data-ustatus]").forEach((b) =>
      b.addEventListener("click", async () => {
        const status = Number(b.dataset.ustatus);
        const action = status === 0 ? "禁用" : "启用";
        if (!confirm("确认" + action + "用户「" + b.dataset.uname + "」？" + (status === 0 ? "（禁用后该用户将立即被踢下线且无法登录）" : ""))) {
          return;
        }
        try {
          await API.adminSetUserStatus(b.dataset.uid, status);
          toast(action + "成功", "success");
          loadAdminUsers();
        } catch (e) {
          toast(e.message, "error");
        }
      })
    );
    $("admin-user-tbody").querySelectorAll("[data-udel]").forEach((b) =>
      b.addEventListener("click", async () => {
        if (!confirm("确认删除用户「" + b.dataset.uname + "」？其仓库、交易记录等数据将一并删除，且不可恢复！")) {
          return;
        }
        try {
          await API.adminDeleteUser(b.dataset.udel);
          toast("删除成功", "success");
          loadAdminUsers();
          loadAdminKeys();
        } catch (e) {
          toast(e.message, "error");
        }
      })
    );

    renderPager("admin-user-pager", state.adminUserPage, totalPages, (p) => {
      state.adminUserPage = p;
      loadAdminUsers();
    });
  } catch (e) {
    toast(e.message, "error");
  }
}

// ---------- 密钥管理 ----------
async function loadAdminKeys() {
  try {
    const keys = await API.adminKeys();
    const fmtTime = (s) => (s || "").replace("T", " ").slice(0, 19);
    $("admin-key-tbody").innerHTML = keys.length
      ? keys
          .map((k) => {
            const remain = k.max_uses - k.used_count;
            const statusHtml =
              k.status !== 1 ? '<span class="trend-down">已禁用</span>'
              : remain <= 0 ? '<span class="trend-flat">已用完</span>'
              : '<span class="trend-up">可用</span>';
            // 已被使用过的密钥后端不允许删除
            const canDel = k.used_count === 0;
            return `<tr>
        <td class="freeze"><code class="key-code">${esc(k.key)}</code></td>
        <td>${k.max_uses}</td>
        <td>${k.used_count}</td>
        <td>${remain}</td>
        <td>${statusHtml}</td>
        <td>${esc(k.created_by || "-")}</td>
        <td>${esc(fmtTime(k.created_at))}</td>
        <td>
          <button class="btn small" data-keyusers="${k.id}" data-key="${esc(k.key)}">查看用户</button>
          ${canDel ? ' <button class="btn small danger" data-keydel="' + k.id + '" data-key="' + esc(k.key) + '">删除</button>' : ""}
        </td>
      </tr>`;
          })
          .join("")
      : '<tr><td class="freeze" colspan="8">（暂无密钥）</td></tr>';

    // 查看密钥绑定用户
    $("admin-key-tbody").querySelectorAll("[data-keyusers]").forEach((b) =>
      b.addEventListener("click", () => showKeyUsers(b.dataset.keyusers, b.dataset.key))
    );
    // 删除密钥（仅未被使用过的）
    $("admin-key-tbody").querySelectorAll("[data-keydel]").forEach((b) =>
      b.addEventListener("click", async () => {
        if (!confirm("确认删除密钥「" + b.dataset.key + "」？删除后该密钥将无法用于注册。")) {
          return;
        }
        try {
          await API.adminDeleteKey(b.dataset.keydel);
          toast("删除成功", "success");
          loadAdminKeys();
        } catch (e) {
          toast(e.message, "error");
        }
      })
    );
  } catch (e) {
    toast(e.message, "error");
  }
}

// 密钥绑定用户弹窗
async function showKeyUsers(id, key) {
  try {
    const users = await API.adminKeyUsers(id);
    $("key-users-title").textContent = "密钥 " + key + " 绑定的用户（" + users.length + " 人）";
    const fmtTime = (s) => (s || "").replace("T", " ").slice(0, 19);
    $("key-users-tbody").innerHTML = users.length
      ? users
          .map(
            (u) => `<tr>
        <td class="freeze">${esc(u.username)}</td>
        <td>第 ${u.day} 天</td>
        <td>${money(u.money)}</td>
        <td>${esc(fmtTime(u.created_at))}</td>
      </tr>`
          )
          .join("")
      : '<tr><td class="freeze" colspan="4">（该密钥暂无绑定用户）</td></tr>';
    $("key-users-modal").classList.remove("hidden");
  } catch (e) {
    toast(e.message, "error");
  }
}

// 新建密钥弹窗
function openKeyModal() {
  $("key-max-uses").value = 10;
  $("key-msg").textContent = "";
  $("key-modal").classList.remove("hidden");
}

async function confirmCreateKey() {
  const maxUses = Number($("key-max-uses").value);
  if (!Number.isInteger(maxUses) || maxUses < 1) {
    $("key-msg").textContent = "可用次数必须是正整数";
    return;
  }
  const btn = $("key-create");
  btn.disabled = true;
  try {
    const k = await API.adminCreateKey(maxUses);
    toast("创建成功：密钥 " + k.key + "（可用 " + k.max_uses + " 次）", "success");
    $("key-modal").classList.add("hidden");
    loadAdminKeys();
  } catch (e) {
    $("key-msg").textContent = e.message;
  } finally {
    btn.disabled = false;
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

// ==================== 数量步进控制（▲/▼ 按步长增减） ====================
function adjustQty(dir) {
  const qtyInput = $("trade-quantity");
  // 步长取输入框值，非法（空/非数字/<1）时按 1 处理
  const stepRaw = Number($("trade-step").value);
  const step = Number.isFinite(stepRaw) && stepRaw >= 1 ? Math.floor(stepRaw) : 1;
  let qty = Math.floor(Number(qtyInput.value) || 0) + dir * step;
  const max = Number(qtyInput.max);
  if (qty < 1) qty = 1;
  if (max > 0 && qty > max) qty = max; // 卖出时不超过持有数量
  qtyInput.value = qty;
  $("trade-estimate").textContent = estimateText();
}

// applyQtyFraction 半仓/全仓快捷填充：
// 买入 → 按剩余仓库空间可存放数量计算（考虑所选存储仓库）；
// 卖出 → 按持有数量计算。
function applyQtyFraction(frac) {
  if (!tradeCtx) return;
  let qty = 0;
  if (tradeCtx.mode === "sell") {
    qty = Math.floor(tradeCtx.product.quantity * frac);
  } else {
    if (!tradeCtx.space) {
      toast("未能获取仓库空间，无法计算", "error");
      return;
    }
    // 当前选择的存储仓库（仅双仓库商品显示选择框）
    const storage = $("trade-storage-label").classList.contains("hidden")
      ? (tradeCtx.product.storage_type === STORAGE_COLD ? STORAGE_COLD : STORAGE_NORMAL)
      : Number($("trade-storage").value);
    const free = storage === STORAGE_COLD ? tradeCtx.space.cold.free : tradeCtx.space.normal.free;
    const size = Number(tradeCtx.product.size) || 1;
    qty = Math.floor(Math.floor(free / size) * frac);
  }
  if (qty < 1) {
    toast(tradeCtx.mode === "sell" ? "持有数量不足" : "剩余仓库空间不足", "error");
    return;
  }
  $("trade-quantity").value = qty;
  $("trade-estimate").textContent = estimateText();
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

async function openBuyModal(p) {
  tradeCtx = { mode: "buy", product: p, space: null };
  // 获取仓库剩余空间（半仓/全仓计算用）
  try {
    tradeCtx.space = await API.warehouseSpace();
  } catch (e) {
    /* 获取失败时点击半仓/全仓会提示 */
  }
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
  // 顶栏"明天"按钮
  $("btn-tomorrow-top").addEventListener("click", doTomorrow);

  // 交易弹窗
  $("trade-quantity").addEventListener("input", () => {
    $("trade-estimate").textContent = estimateText();
  });
  // 数量步进按钮（按步长输入框的值增减）
  $("qty-up").addEventListener("click", () => adjustQty(1));
  $("qty-down").addEventListener("click", () => adjustQty(-1));
  // 半仓 / 全仓快捷填充
  $("qty-half").addEventListener("click", () => applyQtyFraction(0.5));
  $("qty-full").addEventListener("click", () => applyQtyFraction(1));
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

  // 密钥和用户管理（仅 admin 可见）
  $("btn-admin").addEventListener("click", () => switchPage("admin"));
  $("btn-new-key").addEventListener("click", openKeyModal);
  $("key-create").addEventListener("click", confirmCreateKey);
  $("key-cancel").addEventListener("click", () => $("key-modal").classList.add("hidden"));
  $("key-users-close").addEventListener("click", () => $("key-users-modal").classList.add("hidden"));

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
