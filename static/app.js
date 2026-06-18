var selectedEl = null;
var templateName = "";

document.querySelectorAll(".tag").forEach(function(btn) {
	btn.addEventListener("click", function() {
		var tag = this.getAttribute("data-tag");
		var el = document.createElement(tag);
		if (tag === "a") { el.href = "#"; el.textContent = "link"; }
		else if (tag === "img") { el.src = ""; el.alt = "image"; }
		else if (tag === "input") { el.type = "text"; el.placeholder = "input"; }
		else if (tag === "button") { el.textContent = "Button"; }
		else if (tag === "ul") { el.innerHTML = "<li>Item 1</li><li>Item 2</li>"; }
		else if (tag === "li") { el.textContent = "Item"; }
		else if (tag === "tr") { el.innerHTML = "<td>Cell</td>"; }
		else if (tag === "td" || tag === "th") { el.textContent = "Cell"; }
		else if (tag === "textarea") { el.textContent = "text"; }
		else if (tag === "select") { el.innerHTML = "<option>Option 1</option><option>Option 2</option>"; }
		else if (tag === "label") { el.textContent = "Label"; }
		else if (tag === "form") { el.innerHTML = '<input type="text" placeholder="Field"><button>Submit</button>'; }
		else { el.textContent = "<" + tag + "> element</" + tag + ">"; }
		el.classList.add("editable");
		el.addEventListener("click", function(e) { selectElement(this); e.stopPropagation(); });
		if (selectedEl && selectedEl !== document.getElementById("preview")) {
			selectedEl.appendChild(el);
		} else {
			document.getElementById("preview").appendChild(el);
		}
		selectElement(el);
	});
});

function selectElement(el) {
	if (selectedEl) selectedEl.classList.remove("selected");
	selectedEl = el;
	el.classList.add("selected");
	makeEditable(el);
}

function makeEditable(el) {
	var props = document.querySelector(".properties");
	var tag = el.tagName.toLowerCase();
	var html = "<h3>&lt;" + tag + "&gt; Properties</h3>";
	var attrs = ["id", "class", "style", "name", "href", "src", "alt", "placeholder", "type", "value"];
	attrs.forEach(function(attr) {
		var val = el.getAttribute(attr) || "";
		html += '<label>' + attr + ': <input class="prop-input" data-attr="' + attr + '" value="' + val.replace(/"/g, "&quot;") + '"></label>';
	});
	html += '<label>Text content: <input class="prop-input text-content" value="' + (el.textContent || "").replace(/"/g, "&quot;") + '"></label>';
	html += '<button class="button remove-btn" style="margin-top:0.5rem;background:#c62828;">Remove</button>';
	props.innerHTML = html;
	props.querySelectorAll(".prop-input").forEach(function(inp) {
		inp.addEventListener("input", function() {
			var attr = this.getAttribute("data-attr");
			if (attr) {
				if (this.value) el.setAttribute(attr, this.value);
				else el.removeAttribute(attr);
			}
		});
	});
	var textInput = props.querySelector(".text-content");
	if (textInput) {
		textInput.addEventListener("input", function() {
			el.textContent = this.value;
		});
	}
	props.querySelector(".remove-btn").addEventListener("click", function() {
		el.remove();
		selectedEl = null;
		props.innerHTML = "<h3>Properties</h3><p>Click an element in the preview to edit its properties.</p>";
	});
}

document.getElementById("preview").addEventListener("click", function(e) {
	if (e.target === this) {
		if (selectedEl) selectedEl.classList.remove("selected");
		selectedEl = null;
		document.querySelector(".properties").innerHTML = "<h3>Properties</h3><p>Click an element in the preview to edit its properties.</p>";
	}
});

function getHTML() {
	var preview = document.getElementById("preview");
	return preview.innerHTML;
}

document.getElementById("save-btn").addEventListener("click", async function() {
	var name = document.getElementById("template-name").value;
	if (!name) { alert("Enter a template name"); return; }
	var html = getHTML();
	var res = await fetch("/api/templates", {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify({ name: name, html: html })
	});
	if (res.ok) alert("Template saved! Access at /" + name);
	else alert("Save failed");
});

document.getElementById("code-btn").addEventListener("click", function() {
	var html = getHTML();
	document.getElementById("code-output").textContent = html;
	document.getElementById("code-modal").style.display = "flex";
});

document.getElementById("close-modal").addEventListener("click", function() {
	document.getElementById("code-modal").style.display = "none";
});

document.getElementById("export-js-btn").addEventListener("click", async function() {
	var res = await fetch("/api/templates");
	var js = await res.text();
	var blob = new Blob([js], { type: "application/javascript" });
	var a = document.createElement("a");
	a.href = URL.createObjectURL(blob);
	a.download = "templates.js";
	a.click();
});
