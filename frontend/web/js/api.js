// api.js 后端 HTTP 接口封装（服务器地址可全局配置，保存在 localStorage）
"use strict";

const API = {
  get base() {
    return localStorage.getItem("tg_server") || "http://127.0.0.1:8080";
  },
  set base(v) {
    localStorage.setItem("tg_server", v.replace(/\/+$/, ""));
  },
  get token() {
    return localStorage.getItem("tg_token") || "";
  },
  set token(v) {
    if (v) {
      localStorage.setItem("tg_token", v);
    } else {
      localStorage.removeItem("tg_token");
    }
  },

  // request 统一请求：解析后端 {code, message, data} 结构
  async request(method, path, body) {
    let resp;
    try {
      resp = await fetch(this.base + path, {
        method: method,
        headers: {
          "Content-Type": "application/json",
          ...(this.token ? { "X-Token": this.token } : {}),
        },
        body: body === undefined ? undefined : JSON.stringify(body),
      });
    } catch (e) {
      throw new Error("无法连接服务器 " + this.base + "，请检查地址或在设置中修改");
    }
    let env;
    try {
      env = await resp.json();
    } catch (e) {
      throw new Error("响应解析失败（HTTP " + resp.status + "）");
    }
    if (env.code !== 0) {
      // token 失效：清除登录态，回到登录页
      if (resp.status === 401) {
        this.token = "";
      }
      throw new Error(env.message || "请求失败");
    }
    return env.data;
  },

  // 连接检查
  async health() {
    let resp;
    try {
      resp = await fetch(this.base + "/healthz");
    } catch (e) {
      throw new Error("无法连接服务器");
    }
    if (!resp.ok) {
      throw new Error("健康检查返回 " + resp.status);
    }
  },

  // ---------- 认证 ----------
  async login(username, password) {
    const data = await this.request("POST", "/api/login", { username, password });
    this.token = data.token;
    return data.user;
  },

  async register(username, password, key) {
    const data = await this.request("POST", "/api/register", { username, password, key });
    this.token = data.token;
    return data.user;
  },

  async logout() {
    try {
      await this.request("POST", "/api/logout");
    } finally {
      this.token = "";
    }
  },

  me() {
    return this.request("GET", "/api/me");
  },

  // ---------- 首页 ----------
  home() {
    return this.request("GET", "/api/home");
  },

  // ---------- 商城 / 交易 ----------
  products() {
    return this.request("GET", "/api/products");
  },

  buy(productID, quantity, storageType) {
    return this.request("POST", "/api/buy", {
      product_id: Number(productID),
      quantity: Number(quantity),
      storage_type: Number(storageType),
    });
  },

  sell(productID, quantity, storageType) {
    return this.request("POST", "/api/sell", {
      product_id: Number(productID),
      quantity: Number(quantity),
      storage_type: Number(storageType),
    });
  },

  tomorrow() {
    return this.request("POST", "/api/tomorrow");
  },

  // ---------- 仓库 ----------
  warehouse() {
    return this.request("GET", "/api/warehouse");
  },

  warehouseSpace() {
    return this.request("GET", "/api/warehouse/space");
  },

  warehouseBuy(storageType, size, months) {
    return this.request("POST", "/api/warehouse/buy", {
      storage_type: Number(storageType),
      size: Number(size),
      months: Number(months),
    });
  },

  // ---------- 交易记录 ----------
  transactions(page, pageSize) {
    return this.request(
      "GET",
      "/api/transactions?page=" + page + "&page_size=" + pageSize
    );
  },
};

// 存储类型（与后端约定一致）
const STORAGE_NORMAL = 1;
const STORAGE_COLD = 2;
const STORAGE_BOTH = 3;

function storageTypeName(t) {
  switch (t) {
    case STORAGE_NORMAL:
      return "普通仓库";
    case STORAGE_COLD:
      return "冷藏仓库";
    case STORAGE_BOTH:
      return "普通/冷藏均可";
  }
  return "未知";
}
