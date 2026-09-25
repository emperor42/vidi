"use strict";

const test = require("node:test");
const assert = require("node:assert/strict");
const { Vidi } = require("./vidi.js");

class FakeElement {
  constructor(tag) {
    this.tagName = tag.toUpperCase();
    this.children = [];
    this.attributes = Object.create(null);
    this.textContent = "";
    this.innerHTML = "";
    this.className = "";
    this.disabled = false;
  }

  appendChild(child) {
    this.children.push(child);
    return child;
  }

  setAttribute(name, value) {
    this.attributes[name] = String(value);
  }

  getAttribute(name) {
    return this.attributes[name] || null;
  }

  addEventListener() {}
}

function renderValue(value, fields, row) {
  const previousDocument = global.document;
  global.document = {
    createElement(tag) {
      return new FakeElement(tag);
    },
    createTextNode(text) {
      const node = new FakeElement("#text");
      node.textContent = text;
      return node;
    },
  };
  try {
    const view = Object.create(Vidi.prototype);
    view.fields = fields;
    return view._card(row || { value: value }, fields, null);
  } finally {
    if (previousDocument === undefined) {
      delete global.document;
    } else {
      global.document = previousDocument;
    }
  }
}

function fieldValues(card) {
  const definitionList = card.children.find((child) => child.tagName === "DL");
  return definitionList.children.filter((child) => child.tagName === "DD");
}

function fieldValue(card, index) {
  return fieldValues(card)[index || 0];
}

function textContent(node) {
  return node.textContent + node.children.map(textContent).join("");
}

test("VIDI links only values with an allowed URL scheme", () => {
  const allowed = [
    "http://example.test/item",
    "https://example.test/item",
    "mailto:person@example.test",
    "tel:+15551234567",
  ];

  for (const value of allowed) {
    const dd = fieldValue(renderValue(value, ["link"], { link: value }));
    assert.equal(dd.children.length, 1);
    assert.equal(dd.children[0].tagName, "A");
    assert.equal(dd.children[0].getAttribute("href"), value);
  }
});

test("VIDI renders unsafe and relative link values as text", () => {
  const unsafe = [
    "javascript:alert(1)",
    "JaVaScRiPt:alert(1)",
    "data:text/html,<script>alert(1)</script>",
    "//evil.example.test/path",
    "/relative/path",
    "ftp://example.test/file",
  ];

  for (const value of unsafe) {
    const dd = fieldValue(renderValue(value, ["link"], { link: value }));
    assert.equal(dd.children.some((child) => child.tagName === "A"), false);
    assert.match(textContent(dd), /alert|relative|evil|ftp/);
  }
});

test("VIDI applies the same scheme check to a trailing URL field", () => {
  const safeValue = "https://example.test/record";
  const safeDD = fieldValue(
    renderValue(safeValue, ["description", "url"], {
      description: "A record",
      url: safeValue,
    }),
    1
  );
  assert.equal(safeDD.children[0].tagName, "A");

  const unsafeValue = "javascript:alert(1)";
  const unsafeDD = fieldValue(
    renderValue(unsafeValue, ["description", "url"], {
      description: "A record",
      url: unsafeValue,
    }),
    1
  );
  assert.equal(unsafeDD.children.some((child) => child.tagName === "A"), false);
  assert.match(textContent(unsafeDD), /javascript:alert\(1\)/);
});
