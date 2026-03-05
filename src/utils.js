export function parseTime(timeStr) {
  if (!timeStr) return 0;
  let total = 0;
  const hours = timeStr.match(/(\d+(?:\.\d+)?)h/g);
  const mins = timeStr.match(/(\d+(?:\.\d+)?)m/g);
  if (hours) hours.forEach(h => { total += parseFloat(h); });
  if (mins) mins.forEach(m => { total += parseFloat(m) / 60; });
  return total;
}

export function parseDate(dateStr) {
  const parts = dateStr.split('/');
  if (parts.length !== 3) return null;
  return new Date(+parts[2], +parts[0] - 1, +parts[1]);
}

export function getWeekBounds(date) {
  const d = new Date(date);
  const day = d.getDay();
  const monday = new Date(d);
  monday.setDate(d.getDate() - ((day + 6) % 7));
  monday.setHours(0, 0, 0, 0);
  const sunday = new Date(monday);
  sunday.setDate(monday.getDate() + 6);
  sunday.setHours(23, 59, 59, 999);
  return { monday, sunday };
}

export function formatDateShort(d) {
  return `${d.getMonth() + 1}/${d.getDate()}/${d.getFullYear()}`;
}

export function groupByDate(tasks) {
  const map = {};
  for (const t of tasks) {
    if (!map[t.date]) map[t.date] = [];
    map[t.date].push(t);
  }
  return Object.keys(map).sort().reverse().map(key => ({
    key,
    tasks: map[key],
    total: map[key].reduce((sum, t) => sum + parseTime(t.timeSpent), 0),
  }));
}

export function sortTasks(tasks, sortBy, sortDir) {
  if (!sortBy) return tasks;
  const dir = sortDir === 'desc' ? -1 : 1;
  return [...tasks].sort((a, b) => {
    let cmp;
    if (sortBy === 'date') {
      const da = parseDate(a.date);
      const db = parseDate(b.date);
      cmp = (da || 0) - (db || 0);
    } else if (sortBy === 'timeSpent') {
      cmp = parseTime(a.timeSpent) - parseTime(b.timeSpent);
    } else {
      cmp = (a[sortBy] || '').localeCompare(b[sortBy] || '');
    }
    return cmp * dir;
  });
}

export function filterCurrentWeek(tasks) {
  return filterWeekByOffset(tasks, 0);
}

export function filterWeekByOffset(tasks, offset) {
  const ref = new Date();
  ref.setDate(ref.getDate() + offset * 7);
  const { monday, sunday } = getWeekBounds(ref);
  const filtered = tasks.filter(t => {
    const d = parseDate(t.date);
    return d && d >= monday && d <= sunday;
  });
  const total = filtered.reduce((sum, t) => sum + parseTime(t.timeSpent), 0);
  return {
    label: `${formatDateShort(monday)} – ${formatDateShort(sunday)}`,
    tasks: filtered,
    total,
  };
}
