// openly - dashboard interactivity
(function () {
    "use strict";

    var BASE_URL = "";
    var baseUrlMeta = document.querySelector('meta[name="base-url"]');
    if (baseUrlMeta) BASE_URL = baseUrlMeta.content;

    // ── Toast Notifications ─────────────────────────────
    function toast(message, type) {
        type = type || "info";
        var container = document.getElementById("toast-container");
        if (!container) return;

        var el = document.createElement("div");
        el.className = "toast toast-" + type;
        el.textContent = message;
        container.appendChild(el);

        setTimeout(function () {
            el.classList.add("toast-out");
            setTimeout(function () { el.remove(); }, 300);
        }, 3500);
    }

    // ── Copy to Clipboard ───────────────────────────────
    function copyToClipboard(text, btn) {
        navigator.clipboard.writeText(text).then(function () {
            var orig = btn.textContent;
            btn.textContent = "Copied!";
            btn.classList.add("copied");
            setTimeout(function () {
                btn.textContent = orig;
                btn.classList.remove("copied");
            }, 1800);
        }).catch(function () {
            toast("Failed to copy", "error");
        });
    }

    // ── Custom Confirm Dialog ───────────────────────────
    var confirmOverlay = document.getElementById("confirm-overlay");
    var confirmTitle = document.getElementById("confirm-title");
    var confirmDesc = document.getElementById("confirm-desc");
    var confirmOk = document.getElementById("confirm-ok");
    var confirmCancel = document.getElementById("confirm-cancel");
    var confirmCallback = null;

    function showConfirm(title, desc, okLabel, onConfirm) {
        if (!confirmOverlay) return;
        confirmTitle.textContent = title;
        confirmDesc.textContent = desc;
        confirmOk.textContent = okLabel || "Delete";
        confirmCallback = onConfirm;
        confirmOverlay.classList.remove("hidden");
        confirmOverlay.setAttribute("aria-hidden", "false");
        confirmOk.focus();
    }

    function hideConfirm() {
        if (!confirmOverlay) return;
        confirmOverlay.classList.add("hidden");
        confirmOverlay.setAttribute("aria-hidden", "true");
        confirmCallback = null;
    }

    if (confirmOk) {
        confirmOk.addEventListener("click", function () {
            if (confirmCallback) confirmCallback();
            hideConfirm();
        });
    }
    if (confirmCancel) {
        confirmCancel.addEventListener("click", hideConfirm);
    }
    if (confirmOverlay) {
        confirmOverlay.addEventListener("click", function (e) {
            if (e.target === confirmOverlay) hideConfirm();
        });
    }

    // ── Edit Dialog ─────────────────────────────────────
    var editOverlay = document.getElementById("edit-overlay");
    var editInput = document.getElementById("edit-code-input");
    var editSave = document.getElementById("edit-save");
    var editCancel = document.getElementById("edit-cancel");
    var editingId = null;

    function showEdit(id, currentCode) {
        if (!editOverlay) return;
        editingId = id;
        editInput.value = currentCode;
        editOverlay.classList.remove("hidden");
        editOverlay.setAttribute("aria-hidden", "false");
        editInput.focus();
        editInput.select();
    }

    function hideEdit() {
        if (!editOverlay) return;
        editOverlay.classList.add("hidden");
        editOverlay.setAttribute("aria-hidden", "true");
        editingId = null;
    }

    if (editSave) {
        editSave.addEventListener("click", function () {
            if (!editingId) return;
            var newCode = editInput.value.trim();
            if (!newCode) {
                toast("Short code cannot be empty", "error");
                return;
            }

            editSave.disabled = true;
            editSave.textContent = "Saving\u2026";

            fetch("/api/urls/" + editingId, {
                method: "PATCH",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ short_code: newCode })
            })
            .then(function (res) {
                return res.json().then(function (data) {
                    return { ok: res.ok, data: data };
                });
            })
            .then(function (result) {
                if (!result.ok) {
                    toast(result.data.error || "Failed to update", "error");
                    return;
                }

                // Update the card in place
                var card = document.querySelector('.url-card[data-id="' + editingId + '"]');
                if (card) {
                    card.dataset.shortCode = newCode;
                    card.dataset.searchable = newCode + " " + (card.querySelector(".url-original") ? card.querySelector(".url-original").textContent : "");
                    var shortLink = card.querySelector(".url-short");
                    if (shortLink) {
                        shortLink.textContent = "/" + newCode;
                        shortLink.href = result.data.short_url;
                    }
                    var copyBtn = card.querySelector(".btn-copy[data-url]");
                    if (copyBtn) {
                        copyBtn.dataset.url = result.data.short_url;
                    }
                    var editBtn = card.querySelector(".btn-edit");
                    if (editBtn) {
                        editBtn.dataset.code = newCode;
                    }
                }

                toast("Short code updated!", "success");
                hideEdit();
            })
            .catch(function () {
                toast("Network error. Please try again.", "error");
            })
            .finally(function () {
                editSave.disabled = false;
                editSave.textContent = "Save";
            });
        });
    }

    if (editCancel) {
        editCancel.addEventListener("click", hideEdit);
    }
    if (editOverlay) {
        editOverlay.addEventListener("click", function (e) {
            if (e.target === editOverlay) hideEdit();
        });
    }

    // ── Keyboard: Escape closes modals, Enter saves edit ─
    document.addEventListener("keydown", function (e) {
        if (e.key === "Escape") {
            hideConfirm();
            hideEdit();
        }
        if (e.key === "Enter" && editOverlay && !editOverlay.classList.contains("hidden")) {
            e.preventDefault();
            if (editSave) editSave.click();
        }
    });

    // ── Create URL Form ─────────────────────────────────
    var createForm = document.getElementById("create-url-form");
    if (createForm) {
        createForm.addEventListener("submit", function (e) {
            e.preventDefault();

            var urlInput = document.getElementById("url-input");
            var codeInput = document.getElementById("custom-code");
            var btn = document.getElementById("create-btn");
            var btnText = btn.querySelector(".btn-text");
            var btnLoading = btn.querySelector(".btn-loading");
            var resultBox = document.getElementById("create-result");
            var resultUrl = document.getElementById("result-url");

            if (!urlInput.value.trim()) return;

            // Show loading
            btnText.classList.add("hidden");
            btnLoading.classList.remove("hidden");
            btn.disabled = true;

            var formData = new FormData();
            formData.append("url", urlInput.value.trim());
            if (codeInput.value.trim()) {
                formData.append("custom_code", codeInput.value.trim());
            }

            var origUrl = urlInput.value.trim();

            fetch("/api/urls", {
                method: "POST",
                body: formData
            })
            .then(function (res) {
                return res.json().then(function (data) {
                    return { ok: res.ok, data: data };
                });
            })
            .then(function (result) {
                if (!result.ok) {
                    toast(result.data.error || "Something went wrong", "error");
                    return;
                }

                // Show result
                resultUrl.href = result.data.short_url;
                resultUrl.textContent = result.data.short_url;
                resultBox.classList.remove("hidden");

                toast("Link created successfully!", "success");

                // Clear form
                urlInput.value = "";
                codeInput.value = "";

                // Dynamically add the new URL card to the list
                addURLCard(result.data, origUrl);
            })
            .catch(function () {
                toast("Network error. Please try again.", "error");
            })
            .finally(function () {
                btnText.classList.remove("hidden");
                btnLoading.classList.add("hidden");
                btn.disabled = false;
            });
        });
    }

    // ── Dynamically add a new URL card ──────────────────
    function addURLCard(data, originalUrl) {
        var list = document.getElementById("urls-list");

        // If there's an empty state, remove it and create the list
        var emptyState = document.querySelector(".empty-state");
        if (emptyState) {
            emptyState.remove();
        }

        if (!list) {
            var section = document.querySelector(".urls-section");
            if (!section) return;
            list = document.createElement("div");
            list.className = "urls-list";
            list.id = "urls-list";
            section.appendChild(list);
        }

        var shortUrl = data.short_url || (window.location.origin + "/" + data.short_code);
        var truncUrl = originalUrl.length > 80 ? originalUrl.substring(0, 80) + "\u2026" : originalUrl;

        var card = document.createElement("div");
        card.className = "url-card anim-fade-up";
        card.dataset.id = data.id;
        card.dataset.shortCode = data.short_code;
        card.dataset.searchable = data.short_code + " " + originalUrl;
        card.innerHTML = '<div class="url-card-main">' +
            '<div class="url-info">' +
                '<div class="url-short-row"><a href="' + escapeHtml(shortUrl) + '" target="_blank" rel="noopener" class="url-short">/' + escapeHtml(data.short_code) + '</a></div>' +
                '<div class="url-original" title="' + escapeHtml(originalUrl) + '">' + escapeHtml(truncUrl) + '</div>' +
            '</div>' +
            '<div class="url-meta">' +
                '<span class="url-clicks"><span class="clicks-count">0</span> clicks</span>' +
                '<span class="url-date">Just now</span>' +
            '</div>' +
            '<div class="url-actions">' +
                '<button class="btn btn-small btn-copy" data-url="' + escapeHtml(shortUrl) + '" title="Copy short URL">Copy</button>' +
                '<button class="btn btn-small btn-outline btn-edit" data-id="' + data.id + '" data-code="' + escapeHtml(data.short_code) + '" title="Edit short code">Edit</button>' +
                '<button class="btn btn-small btn-outline btn-stats" data-id="' + data.id + '" title="View click statistics">Stats</button>' +
                '<button class="btn btn-small btn-danger btn-delete" data-id="' + data.id + '" title="Delete this link">Delete</button>' +
            '</div>' +
        '</div>' +
        '<div class="url-stats-panel hidden" id="stats-' + data.id + '">' +
            '<div class="stats-loading">Loading analytics\u2026</div>' +
        '</div>';

        list.insertBefore(card, list.firstChild);
        updateStats();
    }

    // ── Result Copy Button ──────────────────────────────
    var resultCopyBtn = document.getElementById("result-copy");
    if (resultCopyBtn) {
        resultCopyBtn.addEventListener("click", function () {
            var url = document.getElementById("result-url").textContent;
            copyToClipboard(url, resultCopyBtn);
        });
    }

    // ── Delegated Click Handlers (URL List) ─────────────
    document.addEventListener("click", function (e) {
        var target = e.target;

        // Copy button
        if (target.classList.contains("btn-copy") && target.dataset.url) {
            copyToClipboard(target.dataset.url, target);
            return;
        }

        // Delete button
        if (target.classList.contains("btn-delete") && target.dataset.id) {
            var id = target.dataset.id;
            var card = target.closest(".url-card");
            var code = card ? (card.dataset.shortCode || "this link") : "this link";

            showConfirm(
                "Delete link",
                "Are you sure you want to delete /" + code + "? This cannot be undone.",
                "Delete",
                function () {
                    fetch("/api/urls/" + id, { method: "DELETE" })
                        .then(function (res) {
                            if (!res.ok) throw new Error("Failed");
                            if (card) {
                                card.style.transition = "opacity 0.3s, transform 0.3s";
                                card.style.opacity = "0";
                                card.style.transform = "translateX(20px)";
                                setTimeout(function () {
                                    card.remove();
                                    updateStats();
                                }, 300);
                            }
                            toast("Link deleted", "success");
                        })
                        .catch(function () {
                            toast("Failed to delete link", "error");
                        });
                }
            );
            return;
        }

        // Edit button
        if (target.classList.contains("btn-edit") && target.dataset.id) {
            showEdit(target.dataset.id, target.dataset.code);
            return;
        }

        // Stats toggle
        if (target.classList.contains("btn-stats") && target.dataset.id) {
            toggleStats(target.dataset.id, target);
            return;
        }
    });

    // ── Stats Panel ─────────────────────────────────────
    function toggleStats(id, btn) {
        var panel = document.getElementById("stats-" + id);
        if (!panel) return;

        if (!panel.classList.contains("hidden")) {
            panel.classList.add("hidden");
            btn.textContent = "Stats";
            return;
        }

        panel.classList.remove("hidden");
        btn.textContent = "Hide";
        panel.innerHTML = '<div class="stats-loading">Loading analytics\u2026</div>';

        fetch("/api/urls/" + id + "/stats")
            .then(function (res) { return res.json(); })
            .then(function (data) {
                renderStats(panel, data);
            })
            .catch(function () {
                panel.innerHTML = '<div class="stats-loading">Failed to load analytics.</div>';
            });
    }

    function renderStats(panel, data) {
        var clicks = data.clicks || [];
        if (clicks.length === 0) {
            panel.innerHTML = '<div class="stats-empty">No clicks recorded yet for this link.</div>';
            return;
        }

        var html = '<div class="stats-grid">';
        html += '<div class="stats-header-row"><span>Time</span><span>Referrer</span><span>IP Address</span></div>';

        for (var i = 0; i < clicks.length; i++) {
            var c = clicks[i];
            var time = formatDate(c.clicked_at);
            var referrer = c.referrer || "\u2014";
            var ip = c.ip_address || "\u2014";

            html += '<div class="click-row">';
            html += '<span class="click-time">' + escapeHtml(time) + '</span>';
            html += '<span class="click-referrer" title="' + escapeHtml(referrer) + '">' + escapeHtml(referrer) + '</span>';
            html += '<span class="click-ip">' + escapeHtml(ip) + '</span>';
            html += '</div>';
        }
        html += '</div>';
        panel.innerHTML = html;
    }

    function formatDate(str) {
        if (!str) return "";
        try {
            var d = new Date(str.replace(" ", "T") + "Z");
            if (isNaN(d.getTime())) return str;
            return d.toLocaleDateString("en-US", {
                month: "short", day: "numeric", year: "numeric",
                hour: "numeric", minute: "2-digit"
            });
        } catch (e) {
            return str;
        }
    }

    function escapeHtml(str) {
        var div = document.createElement("div");
        div.appendChild(document.createTextNode(str));
        return div.innerHTML;
    }

    // ── Search / Filter URLs ────────────────────────────
    var searchInput = document.getElementById("search-urls");
    if (searchInput) {
        searchInput.addEventListener("input", function () {
            var query = searchInput.value.toLowerCase();
            var cards = document.querySelectorAll(".url-card");
            for (var i = 0; i < cards.length; i++) {
                var searchable = (cards[i].dataset.searchable || "").toLowerCase();
                cards[i].style.display = searchable.indexOf(query) !== -1 ? "" : "none";
            }
        });
    }

    // ── Update stats after add/deletion ─────────────────
    function updateStats() {
        var cards = document.querySelectorAll(".url-card");
        var statNumbers = document.querySelectorAll(".stat-number");
        if (statNumbers.length >= 1) {
            statNumbers[0].textContent = cards.length;
        }
    }

    // ── Staggered entrance animation ────────────────────
    var staggerElements = document.querySelectorAll(".url-card");
    for (var i = 0; i < staggerElements.length; i++) {
        staggerElements[i].style.setProperty("--i", i);
        staggerElements[i].classList.add("anim-stagger");
    }

    // ── Secret Lock Login Access ────────────────────────
    // Click the lock emoji 🔒 on the landing page 5 times to go to /login
    var lockIcon = document.getElementById("lock-icon");
    if (lockIcon) {
        var lockClicks = 0;
        var lockTimer = null;

        lockIcon.style.cursor = "pointer";
        lockIcon.addEventListener("click", function (e) {
            e.stopPropagation();
            lockClicks++;

            if (lockTimer) clearTimeout(lockTimer);

            if (lockClicks >= 5) {
                lockClicks = 0;
                window.location.href = "/login";
                return;
            }

            // Reset after 2 seconds of no clicks
            lockTimer = setTimeout(function () {
                lockClicks = 0;
            }, 2000);
        });
    }
})();
