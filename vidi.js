"use strict";

/**
 * vidi — render database output coming from pod.
 *
 * Fetches records from a pod endpoint (or any JSON endpoint with a compatible
 * shape: { records: […] } or { count, records: [{ id, fields: {…} }] }) and
 * renders them as a paged set of cards, with HTML escaping on every value, a
 * working pager, and CSV export.
 *
 * Usage:
 *   new Vidi({ dataSource: "/api/pod/table/acme/items?format=json", pageSize: 12 });
 *   // renders into #vidi-cards-container and #vidi-pagination by default
 *
 * Browser global: window.Vidi (class) and window.vidi (same class, for the
 * `new Vidi({…})` contract). No external dependencies.
 */

(function () {
  "use strict";

  var ESC = {
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
  };

  function esc(s) {
    return String(s == null ? "" : s).replace(/[&<>"']/g, function (c) {
      return ESC[c];
    });
  }

  function el(tag, attrs, children) {
    var n = document.createElement(tag);
    for (var k in attrs || {}) {
      if (Object.prototype.hasOwnProperty.call(attrs, k)) {
        var v = attrs[k];
        if (k === "disabled") {
          n.disabled = !!v;
        } else if (v == null || v === "") {
          continue; // never setAttribute(null) — that would set the attribute
        } else if (k === "class") n.className = v;
        else if (k === "text") n.textContent = v;
        else if (k === "html") n.innerHTML = v; // only used internally with escaped/trusted content
        else n.setAttribute(k, v);
      }
    }
    (children || []).forEach(function (c) {
      if (c == null) return;
      n.appendChild(typeof c === "string" ? document.createTextNode(c) : c);
    });
    return n;
  }

  // Normalize a record into a flat object of {field: value} pairs. pod rows
  // arrive as { id, fields: {…} }; other sources may send flat maps. Any
  // nested container is flattened so cards can show everything.
  function normalizeRow(rec) {
    var row = {};
    if (!rec) return row;
    if (rec.fields && typeof rec.fields === "object" && !Array.isArray(rec.fields)) {
      for (var f in rec.fields) {
        if (Object.prototype.hasOwnProperty.call(rec.fields, f)) row[f] = rec.fields[f];
      }
    } else if (rec.value && typeof rec.value === "object") {
      for (var v in rec.value) {
        if (Object.prototype.hasOwnProperty.call(rec.value, v)) row[v] = rec.value[v];
      }
    }
    for (var k in rec) {
      if (k === "fields" || k === "value") continue;
      if (typeof rec[k] === "object" && rec[k] !== null) continue;
      if (row[k] === undefined) row[k] = rec[k];
    }
    // Always ensure a visible identifier.
    if (row.id === undefined && rec.id !== undefined) row.id = rec.id;
    return row;
  }

  function prettyField(name) {
    return String(name).replace(/[_-]+/g, " ").replace(/\b\w/g, function (c) {
      return c.toUpperCase();
    });
  }

  class Vidi {
    /**
     * @param {Object} opts
     *   dataSource  URL of the pod/JSON endpoint (required)
     *   container   CSS selector for the cards container (default #vidi-cards-container)
     *   pagination  CSS selector for the pager (default #vidi-pagination)
     *   pageSize    records per page (default 12)
     *   fields      optional array of field names to display (default: all but bookkeeping)
     *   titleField  optional field used as the card title (default: title if present)
     *   onRender    optional callback after each render (rows, container)
     */
    constructor(opts) {
      opts = opts || {};
      this.dataSource = opts.dataSource;
      if (!this.dataSource) throw new Error("vidi: dataSource is required");
      this.pageSize = opts.pageSize || 12;
      this.page = 0;
      this.rows = [];
      this.fields = opts.fields || null;
      this.titleField = opts.titleField || null;
      this.onRender = opts.onRender || null;
      this.container = document.querySelector(opts.container || "#vidi-cards-container");
      this.pagination = document.querySelector(opts.pagination || "#vidi-pagination");
      this.escaped = true; // always escape; kept as a flag for introspection
      this.loading = false;
      this.load();
    }

    load() {
      if (this.loading) return;
      this.loading = true;
      var self = this;
      if (this.container) {
        this.container.innerHTML = "";
        this.container.appendChild(el("div", { class: "vidi-loading" }, ["Loading…"]));
      }
      fetch(this.dataSource, {
        credentials: "same-origin",
        headers: { Accept: "application/json, application/xml, text/xml, */*" },
      })
        .then(function (res) {
          if (!res.ok) throw new Error("HTTP " + res.status);
          var ct = (res.headers.get("content-type") || "").toLowerCase();
          if (ct.indexOf("xml") !== -1) {
            return res.text().then(function (txt) { return self._parseXML(txt); });
          }
          return res.json();
        })
        .then(function (data) {
          self.loading = false;
          self._ingest(data);
        })
        .catch(function (err) {
          self.loading = false;
          self._error(err);
        });
    }

    // pod can return XML; parse it into the same record shape. The XML layout
    // pod emits is <records><record id="…"><field name="…">value</field>…
    // </record></records>.
    _parseXML(xmlText) {
      var doc = new DOMParser().parseFromString(xmlText, "text/xml");
      var root = doc.documentElement;
      if (!root || root.nodeName === "parsererror") {
        throw new Error("vidi: unparseable XML response");
      }
      var out = [];
      var recNodes = root.getElementsByTagName("record");
      for (var i = 0; i < recNodes.length; i++) {
        var rn = recNodes[i];
        var row = { id: rn.getAttribute("id") || "" };
        var fs = rn.getElementsByTagName("field");
        for (var j = 0; j < fs.length; j++) {
          var fn = fs[j].getAttribute("name") || "field" + j;
          row[fn] = fs[j].textContent || "";
        }
        out.push(row);
      }
      return { records: out };
    }

    _ingest(data) {
      // Accept several shapes: {records:[…]}, {items:[…]}, or a bare array.
      var records = null;
      if (Array.isArray(data)) records = data;
      else if (data && Array.isArray(data.records)) records = data.records;
      else if (data && Array.isArray(data.items)) records = data.items;
      else if (data && data.record) records = [data.record];
      if (!records) {
        this._error(new Error("vidi: no records in response"));
        return;
      }
      this.rows = records.map(normalizeRow);
      this.total = this.rows.length;
      this.page = 0;
      this._render();
    }

    _selectedFields() {
      if (this.fields && this.fields.length) return this.fields.slice();
      var names = [];
      var seen = {};
      var skip = { id: 1, created: 1, updated: 1, version: 1, schema_version: 1 };
      for (var i = 0; i < this.rows.length; i++) {
        for (var k in this.rows[i]) {
          if (!Object.prototype.hasOwnProperty.call(this.rows[i], k)) continue;
          if (skip[k]) continue;
          if (!seen[k]) {
            seen[k] = true;
            names.push(k);
          }
        }
      }
      // Deterministic order, id first.
      names.sort();
      return names;
    }

    _render() {
      var fields = this._selectedFields();
      var totalPages = Math.max(1, Math.ceil(this.rows.length / this.pageSize));
      if (this.page >= totalPages) this.page = totalPages - 1;
      if (this.page < 0) this.page = 0;
      var start = this.page * this.pageSize;
      var slice = this.rows.slice(start, start + this.pageSize);

      var self = this;
      if (this.container) {
        this.container.innerHTML = "";
        if (slice.length === 0) {
          this.container.appendChild(el("div", { class: "vidi-empty" }, ["No records"]));
        } else {
          slice.forEach(function (row) {
            self.container.appendChild(self._card(row, fields, self._titleField(row)));
          });
        }
      }
      this._renderPager(totalPages);
      if (this.onRender) {
        try { this.onRender(this.rows.slice(), this.container); } catch (e) { /* callback errors are non-fatal */ }
      }
    }

    _titleField(row) {
      if (this.titleField && row[this.titleField] !== undefined) return row[this.titleField];
      var preferred = ["title", "name", "subject", "label"];
      for (var i = 0; i < preferred.length; i++) {
        var v = row[preferred[i]];
        if (v !== undefined && v !== null && String(v) !== "") return v;
      }
      return null;
    }

    _card(row, fields, title) {
      var card = el("article", { class: "vidi-card vidi-row" });
      var head = el("header", { class: "vidi-card-head" });
      if (title != null) {
        head.appendChild(el("h3", { class: "vidi-card-title", text: String(title) }));
      } else if (row.id !== undefined && row.id !== "") {
        head.appendChild(el("h3", { class: "vidi-card-title", text: String(row.id) }));
      }
      card.appendChild(head);

      var dl = el("dl", { class: "vidi-fields" });
      fields.forEach(function (f) {
        var v = row[f];
        if (v === undefined || v === null) return;
        var s = String(v);
        if (s === "") return;
        dl.appendChild(el("dt", { class: "vidi-field-name" }, [prettyField(f)]));
        var dd = el("dd", { class: "vidi-field-value" });
        if (f === "link" || (looksLikeURL(s) && fields.indexOf(f) === fields.length - 1)) {
          dd.appendChild(el("a", { href: s, target: "_blank", rel: "noopener noreferrer", text: s }));
        } else {
          // Safe: every dynamic value goes through esc() before innerHTML.
          dd.innerHTML = esc(s).replace(/\n/g, "<br>");
        }
        dl.appendChild(dd);
      });
      card.appendChild(dl);
      return card;
    }

    _renderPager(totalPages) {
      var self = this;
      if (!this.pagination) return;
      this.pagination.innerHTML = "";
      if (totalPages <= 1 && this.rows.length === 0) return;

      var nav = el("nav", { class: "vidi-pager", "aria-label": "Pagination" });
      var count = el("span", { class: "vidi-count" }, [
        this.rows.length + " records · page " + (this.page + 1) + " of " + totalPages,
      ]);
      nav.appendChild(count);

      function btn(label, go, disabled) {
        var b = el("button", {
          class: "vidi-page-btn" + (go === self.page ? " active" : ""),
          type: "button",
          disabled: disabled ? "disabled" : null,
        }, [label]);
        b.addEventListener("click", function () {
          if (go < 0 || go >= totalPages || go === self.page) return;
          self.page = go;
          self._render();
          self._scrollToTop();
        });
        return b;
      }

      nav.appendChild(btn("‹ Prev", self.page - 1, self.page === 0));
      // Window of page buttons around the current page.
      var from = Math.max(0, self.page - 2);
      var to = Math.min(totalPages - 1, from + 4);
      from = Math.max(0, to - 4);
      for (var p = from; p <= to; p++) nav.appendChild(btn(String(p + 1), p, false));
      nav.appendChild(btn("Next ›", self.page + 1, self.page >= totalPages - 1));

      // CSV export is always available.
      var csv = el("button", { class: "vidi-page-btn vidi-csv", type: "button" }, ["Export CSV"]);
      csv.addEventListener("click", function () { self.exportCSV(); });
      nav.appendChild(csv);

      this.pagination.appendChild(nav);
    }

    _scrollToTop() {
      if (this.container && this.container.scrollIntoView) {
        try { this.container.scrollIntoView({ block: "start", behavior: "smooth" }); } catch (e) { /* noop */ }
      }
    }

    _error(err) {
      if (this.container) {
        this.container.innerHTML = "";
        this.container.appendChild(el("div", { class: "vidi-error" }, [
          "vidi: " + (err && err.message ? err.message : "failed to load data"),
        ]));
      }
      if (this.onError) this.onError(err);
    }

    /** Export the currently loaded rows as a CSV file (all pages). */
    exportCSV() {
      var fields = this._selectedFields();
      var lines = [];
      lines.push(fields.map(function (f) { return csvCell(prettyField(f)); }).join(","));
      this.rows.forEach(function (row) {
        lines.push(fields.map(function (f) { return csvCell(row[f]); }).join(","));
      });
      var csvData = "\uFEFF" + lines.join("\r\n");
      var blob = new Blob([csvData], { type: "text/csv;charset=utf-8" });
      var url = URL.createObjectURL(blob);
      var a = document.createElement("a");
      a.href = url;
      a.download = "pod-export.csv";
      document.body.appendChild(a);
      a.click();
      setTimeout(function () {
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
      }, 0);
      return lines;
    }
  }

  function csvCell(v) {
    var s = String(v == null ? "" : v);
    if (s.indexOf(",") !== -1 || s.indexOf('"') !== -1 || s.indexOf("\n") !== -1) {
      return '"' + s.replace(/"/g, '""') + '"';
    }
    return s;
  }

  function looksLikeURL(s) {
    return /^(https?|mailto|tel):\/?\/?/i.test(s);
  }

  // Auto-init readiness: the contract is `new Vidi({dataSource})`, so no
  // automatic fetch happens here — callers construct it explicitly.

  // Export for module systems
  if (typeof module !== "undefined" && module.exports) {
    module.exports = { Vidi: Vidi };
  }

  if (typeof window !== "undefined") {
    window.Vidi = Vidi;
    window.vidi = Vidi;
  }
})();