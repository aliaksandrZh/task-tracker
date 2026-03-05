import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { parseTime, parseDate, getWeekBounds, formatDateShort, groupByDate, filterWeekByOffset, sortTasks } from '../src/utils.js';

describe('parseTime', () => {
  it('returns 0 for empty string', () => {
    assert.equal(parseTime(''), 0);
  });

  it('returns 0 for null/undefined', () => {
    assert.equal(parseTime(null), 0);
    assert.equal(parseTime(undefined), 0);
  });

  it('parses hours', () => {
    assert.equal(parseTime('2h'), 2);
  });

  it('parses minutes', () => {
    assert.equal(parseTime('30m'), 0.5);
  });

  it('parses compound time', () => {
    assert.equal(parseTime('1h 30m'), 1.5);
  });

  it('parses decimal hours', () => {
    assert.equal(parseTime('1.5h'), 1.5);
  });

  it('parses decimal compound', () => {
    assert.equal(parseTime('1.5h 15m'), 1.75);
  });
});

describe('parseDate', () => {
  it('parses M/D/YYYY', () => {
    const d = parseDate('3/3/2026');
    assert.equal(d.getFullYear(), 2026);
    assert.equal(d.getMonth(), 2);
    assert.equal(d.getDate(), 3);
  });

  it('parses MM/DD/YYYY', () => {
    const d = parseDate('12/25/2026');
    assert.equal(d.getMonth(), 11);
    assert.equal(d.getDate(), 25);
  });

  it('returns null for invalid format', () => {
    assert.equal(parseDate('2026-03-03'), null);
    assert.equal(parseDate('not a date'), null);
    assert.equal(parseDate(''), null);
  });
});

describe('getWeekBounds', () => {
  it('returns monday to sunday for a wednesday', () => {
    const { monday, sunday } = getWeekBounds(new Date(2026, 2, 4)); // Wed Mar 4
    assert.equal(monday.getDay(), 1);
    assert.equal(sunday.getDay(), 0);
    assert.equal(monday.getDate(), 2);
    assert.equal(sunday.getDate(), 8);
  });

  it('returns same week when given a monday', () => {
    const { monday, sunday } = getWeekBounds(new Date(2026, 2, 2)); // Mon Mar 2
    assert.equal(monday.getDate(), 2);
    assert.equal(sunday.getDate(), 8);
  });

  it('returns same week when given a sunday', () => {
    const { monday, sunday } = getWeekBounds(new Date(2026, 2, 8)); // Sun Mar 8
    assert.equal(monday.getDate(), 2);
    assert.equal(sunday.getDate(), 8);
  });
});

describe('formatDateShort', () => {
  it('formats date as M/D/YYYY', () => {
    assert.equal(formatDateShort(new Date(2026, 2, 3)), '3/3/2026');
  });

  it('does not pad single digits', () => {
    assert.equal(formatDateShort(new Date(2026, 0, 5)), '1/5/2026');
  });

  it('formats double-digit month/day', () => {
    assert.equal(formatDateShort(new Date(2026, 11, 25)), '12/25/2026');
  });
});

describe('groupByDate', () => {
  const tasks = [
    { date: '3/1/2026', type: 'Bug', number: '1', name: 'A', timeSpent: '1h', comments: '' },
    { date: '3/2/2026', type: 'Task', number: '2', name: 'B', timeSpent: '2h', comments: '' },
    { date: '3/1/2026', type: 'Bug', number: '3', name: 'C', timeSpent: '30m', comments: '' },
    { date: '3/3/2026', type: 'Task', number: '4', name: 'D', timeSpent: '', comments: '' },
  ];

  it('groups tasks by date', () => {
    const groups = groupByDate(tasks);
    assert.equal(groups.length, 3);
  });

  it('sorts most recent first', () => {
    const groups = groupByDate(tasks);
    assert.equal(groups[0].key, '3/3/2026');
    assert.equal(groups[1].key, '3/2/2026');
    assert.equal(groups[2].key, '3/1/2026');
  });

  it('calculates totals per group', () => {
    const groups = groupByDate(tasks);
    const mar1 = groups.find(g => g.key === '3/1/2026');
    assert.equal(mar1.total, 1.5);
    assert.equal(mar1.tasks.length, 2);
  });

  it('handles empty time as 0', () => {
    const groups = groupByDate(tasks);
    const mar3 = groups.find(g => g.key === '3/3/2026');
    assert.equal(mar3.total, 0);
  });

  it('returns empty array for no tasks', () => {
    assert.deepEqual(groupByDate([]), []);
  });
});

describe('sortTasks', () => {
  const tasks = [
    { date: '3/3/2026', type: 'Task', number: '200', name: 'Beta', timeSpent: '2h', comments: '' },
    { date: '3/1/2026', type: 'Bug', number: '100', name: 'Alpha', timeSpent: '30m', comments: '' },
    { date: '3/2/2026', type: 'Task', number: '150', name: 'Gamma', timeSpent: '1h', comments: '' },
  ];

  it('returns tasks unchanged when sortBy is null', () => {
    const result = sortTasks(tasks, null, 'asc');
    assert.deepEqual(result, tasks);
  });

  it('sorts by date ascending', () => {
    const result = sortTasks(tasks, 'date', 'asc');
    assert.equal(result[0].date, '3/1/2026');
    assert.equal(result[1].date, '3/2/2026');
    assert.equal(result[2].date, '3/3/2026');
  });

  it('sorts by date descending', () => {
    const result = sortTasks(tasks, 'date', 'desc');
    assert.equal(result[0].date, '3/3/2026');
    assert.equal(result[2].date, '3/1/2026');
  });

  it('sorts by type alphabetically', () => {
    const result = sortTasks(tasks, 'type', 'asc');
    assert.equal(result[0].type, 'Bug');
    assert.equal(result[1].type, 'Task');
  });

  it('sorts by number as string', () => {
    const result = sortTasks(tasks, 'number', 'asc');
    assert.equal(result[0].number, '100');
    assert.equal(result[1].number, '150');
    assert.equal(result[2].number, '200');
  });

  it('sorts by name alphabetically', () => {
    const result = sortTasks(tasks, 'name', 'asc');
    assert.equal(result[0].name, 'Alpha');
    assert.equal(result[1].name, 'Beta');
    assert.equal(result[2].name, 'Gamma');
  });

  it('sorts by timeSpent numerically', () => {
    const result = sortTasks(tasks, 'timeSpent', 'asc');
    assert.equal(result[0].timeSpent, '30m');
    assert.equal(result[1].timeSpent, '1h');
    assert.equal(result[2].timeSpent, '2h');
  });

  it('sorts by timeSpent descending', () => {
    const result = sortTasks(tasks, 'timeSpent', 'desc');
    assert.equal(result[0].timeSpent, '2h');
    assert.equal(result[2].timeSpent, '30m');
  });

  it('does not mutate the original array', () => {
    const original = [...tasks];
    sortTasks(tasks, 'name', 'asc');
    assert.deepEqual(tasks, original);
  });
});

describe('filterWeekByOffset', () => {
  // Create tasks spanning multiple weeks
  const tasks = [
    { date: '3/2/2026', type: 'Bug', number: '1', name: 'A', timeSpent: '1h', comments: '' },
    { date: '3/3/2026', type: 'Task', number: '2', name: 'B', timeSpent: '2h', comments: '' },
    { date: '2/23/2026', type: 'Bug', number: '3', name: 'C', timeSpent: '30m', comments: '' },
  ];

  it('offset 0 returns current week (same as filterCurrentWeek)', () => {
    const result = filterWeekByOffset(tasks, 0);
    assert.ok(result.label);
    assert.ok(Array.isArray(result.tasks));
    assert.equal(typeof result.total, 'number');
  });

  it('negative offset returns earlier week', () => {
    const current = filterWeekByOffset(tasks, 0);
    const prev = filterWeekByOffset(tasks, -1);
    // Labels should be different
    assert.notEqual(current.label, prev.label);
  });

  it('returns empty tasks for week with no data', () => {
    const farPast = filterWeekByOffset(tasks, -100);
    assert.equal(farPast.tasks.length, 0);
    assert.equal(farPast.total, 0);
  });

  it('label contains date range', () => {
    const result = filterWeekByOffset(tasks, 0);
    assert.ok(result.label.includes('–'));
  });
});
