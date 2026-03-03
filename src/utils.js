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

export function filterCurrentWeek(tasks) {
  const { monday, sunday } = getWeekBounds(new Date());
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
