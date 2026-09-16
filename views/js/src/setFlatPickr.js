/*
 * Copyright (c) 2026. Author: Ruslan Bikchentaev. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 * Перший приватний програміст. 
 */
"use strict";

const DATES_CONSTRAINTS_KEY = 'CONSTRAINTS_DATE';

class FlatPickrManager {
    constructor() {
        // month-select range state: null between clicks, the start-of-month
        // Date while the first half of a two-click range pick is pending.
        this._monthRangeStart = null;
    }

    // Entry point used by set_clicks.js: prepare every [type=date] and
    // [type=date-range] input under `parent` that hasn't been handled yet.
    init(parent) {
        this._setupType('date-range', parent);
        this._setupType('date', parent);
}

    _setupType(typ, parent) {
    let dates = $(`input[type=${typ}]:not([rel])`, parent);
    if (dates.length === 0) {
        return;
    }

    if (window.flatpickr !== undefined) {
        this._apply(dates, typ);
        return;
    }

    LoadJScript("https://cdn.jsdelivr.net/npm/flatpickr", false, true, () => {
        if (dates.filter('[plugin="month"]').length > 0 && window.monthSelectPlugin === undefined) {
            loadCSSOnce("https://cdn.jsdelivr.net/npm/flatpickr/dist/plugins/monthSelect/style.css");
            LoadJScript("https://cdn.jsdelivr.net/npm/flatpickr/dist/plugins/monthSelect/index.js", true, true,
                () => {
                    this._apply(dates, typ);
                });
            return;
        }
        this._apply(dates, typ);
    });
}

    _readConstraints() {
        return JSON.parse(localStorage.getItem(DATES_CONSTRAINTS_KEY)) || {minDate: 1433141722, maxDate: new Date()};
    }

    _apply(dates, typ) {
    const isSingle = typ === 'date';
        const selDates = this._readConstraints();
        const ymd = "Y-m-d";

    dates.each((id, elem) => {
        const isMonth = elem.getAttribute("plugin") === "month";

        let cfg = {
            animate: true,
            altInput: true, // Allows a different display format
            locale: {
                firstDayOfWeek: 1 // Monday
                // document.documentElement.lang.split(",")[0] || 'en',
            },
            mode: isSingle || isMonth ? 'single' : 'range',
            // minDate: UnixTimetoDate(selDates['minDate']),
            maxDate: elem.max || selDates['maxDate'],
            onReady: (selectedDates, dateStr, instance) => {
                this._attachWeekClickHandler(instance);
                const val = localStorage.getItem(elem.id);
                instance._input.value = val;
            },
        };

        cfg = isMonth ? this._monthConfig(cfg, isSingle) : this._rangeConfig(cfg, isSingle, ymd, selDates);

        elem.flatpickr(cfg);
        elem.rel = typ;
    });
    }

    _monthConfig(cfg, isSingle) {
            const monthFormat = "Y-m-d";
            const monthRange = `[${monthFormat},${monthFormat}]`;

        return {
                ...cfg,
                mode: "single",
                plugins: [
                    new monthSelectPlugin({
                        shorthand: true, //defaults to false
                        dateFormat: isSingle ? monthFormat : monthRange, //defaults to "F Y"
                        altFormat: isSingle ? "F Y" : monthRange, //defaults to "F Y"
                        // theme: "dark" // defaults to "light"
                    })
                ],
                // IMPORTANT: allow manual control
                clickOpens: true,
                closeOnSelect: false,
            onClose: (dates, dateStr, instance) => {
                    localStorage.setItem(instance.element.id, instance.input.value);
                },
            onChange: (selectedDates, _, instance) => {
                    if (!selectedDates.length) return;

                    const d = selectedDates[0];
                    const monthStart = new Date(d.getFullYear(), d.getMonth(), 1);
                    const monthEnd = new Date(d.getFullYear(), d.getMonth() + 1, 0);

                    // ---- first click ----
                if (!this._monthRangeStart) {
                    this._monthRangeStart = monthStart;
                        instance.input.value = `[${instance.formatDate(monthStart, monthFormat)}, …]"`;
                        return;
                    }

                    // ---- second click ----
                const start = this._monthRangeStart < monthStart ? this._monthRangeStart : monthStart;
                const end = this._monthRangeStart < monthStart ? monthEnd : this._monthRangeStart;

                    instance.input.value = `[${instance.formatDate(start, monthFormat)},${instance.formatDate(end, monthFormat)}]`;
                this._monthRangeStart = null;
                    instance.close();
                }
            };
    }

    _rangeConfig(cfg, isSingle, ymd, selDates) {
        return {
                ...cfg,
                altInput: true, // Allows a different display format
                altFormat:isSingle ? ymd : "[Y-m-d,Y-m-d]",
                inputFormat:isSingle ? ymd : "[Y-m-d,Y-m-d]",
                dateFormat: isSingle ? ymd : "[Y-m-d,Y-m-d]", // Flatpickr saves this format
                disabled: selDates['holidays'],
                weekNumbers: true,
            onClose: (dates, dateStr, instance) => {
                    if (dates.length === 2) {
                        // Format value as [YYYY-MM-DD,YYYY-MM-DD]
                        instance.input.value = `[${instance.formatDate(dates[0], ymd)},${instance.formatDate(dates[1], ymd)}]`;
                    } else {
                        // If only one date is selected, clear the input
                        instance.input.value = dateStr;
                        console.log(dateStr);
                    }
                    localStorage.setItem(instance.element.id, instance.input.value);
                },
            onMonthChange: (dates, dateStr, instance) => {
                    const y = instance.currentYear;
                    const m = instance.currentMonth; // still 0-based

                    const first = new Date(y, m, 1);
                    const last = new Date(y, m + 1, 0);
                    const maxDate = instance.config.maxDate;

                    instance.setDate([first, last > maxDate ? maxDate : last], true);
                this._setWeekStyles(instance.calendarContainer);
                    submitWhenAlt(instance);
                },
            onOpen: (dates, dateStr, instance) => {
                this._setWeekStyles(instance.calendarContainer);
                },
            };
        }

    _setWeekStyles(container) {
    container.querySelectorAll(
        ".flatpickr-weekwrapper .flatpickr-weeks > .flatpickr-day"
    ).forEach(weekElem => {
        // 💡 Tooltip
        weekElem.title = "Click to select whole week (+Alt run GENERATE)";
        weekElem.style.cursor = "pointer";
    });
}

    _attachWeekClickHandler(instance) {
    if (!instance.config.weekNumbers) {
        return;
    }
    const container = instance.calendarContainer;

    // prevent double-attaching the delegator
    if (container.dataset.weekDelegatorAttached) return;
    container.dataset.weekDelegatorAttached = "1";
        this._setWeekStyles(container);

    container.addEventListener("click", (evt) => {
        const weekElem = evt.target.closest(".flatpickr-weekwrapper .flatpickr-weeks .flatpickr-day");
        // NOTE: prefer ".flatpickr-week" — adjust selector if your flatpickr version uses a different class.
        if (!weekElem) return;

        // Some flatpickr versions render week number as a node inside .flatpickr-week; try to extract numeric content robustly
        const txt = (weekElem.textContent || "").trim();
        const weekNumber = parseInt(txt, 10);
        if (!Number.isFinite(weekNumber)) return;

        // Get year currently visible
        const year = instance.currentYear;

        let startDate = this.isoWeekStart(year, weekNumber);
        let endDate = new Date(startDate);

        endDate.setDate(endDate.getDate() + 6); // full 7-day span
        if (endDate > instance.config.maxDate) {
            endDate = instance.config.maxDate;
        }

// --- SHIFT: expand previous range ---
        const keys = window.keyState || {};
        if (keys.shift && instance.selectedDates.length > 0) {
            const prevSelect = instance.selectedDates;

            if (prevSelect[0] < startDate) {
                startDate = prevSelect[0];
            }
            if (prevSelect[1] > endDate) {
                endDate = prevSelect[1]
            }
        }
        // Apply
        instance.setDate([startDate, endDate], true);

        // --- CMD key logic (optional) ---
        if (keys.meta) {
            console.log("Cmd pressed — custom logic here");
        }
        submitWhenAlt(instance);
    });
}

    isoWeekStart(year, week) {
    const simple = new Date(year, 0, 1 + (week - 1) * 7);
    const dow = simple.getDay();
    const ISOweekStart = simple;
    if (dow <= 4)
        ISOweekStart.setDate(simple.getDate() - simple.getDay() + 1);
    else
        ISOweekStart.setDate(simple.getDate() + 8 - simple.getDay());
    return ISOweekStart;
}

    selectDay() {
        const elem = htmx.find('input[name=calendar]')._flatpickr;
        elem.setDate([elem.config.maxDate]);
        submitAfterUpdate(elem._input);
        return false;
    }

    selectWeek() {
        const elem = htmx.find('input[name=calendar]')._flatpickr;
        const endDate = elem.config.maxDate;
        let startDate = new Date(endDate);
        startDate.setDate(startDate.getDate() - endDate.getDay() + 1)
        elem.setDate([startDate, endDate]);
        elem.close();
        submitAfterUpdate(elem._input);
        return false;
    }
}

const flatPickrManager = new FlatPickrManager();

// --- Backward-compatible global function names ---
function setFlatPickr(parent) {
    flatPickrManager.init(parent);
}

function SelectDay() {
    return flatPickrManager.selectDay();
}

function SelectWeek(currentWeek) {
    return flatPickrManager.selectWeek(currentWeek);
}

function isoWeekStart(year, week) {
    return flatPickrManager.isoWeekStart(year, week);
}

function getThisWeekDates() {
    return ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
}