import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { parsePastedText } from '../src/parser.js';

describe('parsePastedText', () => {
  it('parses full format with date, type, number, name, time', () => {
    const [t] = parsePastedText('3/3/2026\nBug 100: Fix crash 2h');
    assert.equal(t.date, '3/3/2026');
    assert.equal(t.type, 'Bug');
    assert.equal(t.number, '100');
    assert.equal(t.name, 'Fix crash');
    assert.equal(t.timeSpent, '2h');
  });

  it('parses type and name without time', () => {
    const [t] = parsePastedText('3/3/2026\nTask 27982: [UI/UX Redesign] Hover Over Tooltip');
    assert.equal(t.type, 'Task');
    assert.equal(t.number, '27982');
    assert.equal(t.name, '[UI/UX Redesign] Hover Over Tooltip');
    assert.equal(t.timeSpent, '');
    assert.ok(t.missing.includes('timeSpent'));
  });

  it('extracts number and name from unknown type word', () => {
    const [t] = parsePastedText('3/3/2026\nask 31557: [Corporate Tags: API] Autocomplete method');
    assert.equal(t.type, '');
    assert.equal(t.number, '31557');
    assert.equal(t.name, '[Corporate Tags: API] Autocomplete method');
    assert.ok(t.missing.includes('type'));
  });

  it('parses number: name without type word', () => {
    const [t] = parsePastedText('3/3/2026\n31557: Some task name');
    assert.equal(t.number, '31557');
    assert.equal(t.name, 'Some task name');
    assert.ok(t.missing.includes('type'));
  });

  it('treats bare text as name only', () => {
    const [t] = parsePastedText('3/3/2026\nJust a task description');
    assert.equal(t.name, 'Just a task description');
    assert.ok(t.missing.includes('type'));
    assert.ok(t.missing.includes('number'));
  });

  it('defaults to today when no date line', () => {
    const [t] = parsePastedText('Bug 1: Test 1h');
    const d = new Date();
    const today = `${d.getMonth() + 1}/${d.getDate()}/${d.getFullYear()}`;
    assert.equal(t.date, today);
    assert.ok(t.missing.includes('date'));
  });

  it('switches date mid-block', () => {
    const tasks = parsePastedText('3/1/2026\nBug 1: A 1h\n3/2/2026\nBug 2: B 2h');
    assert.equal(tasks[0].date, '3/1/2026');
    assert.equal(tasks[1].date, '3/2/2026');
  });

  it('parses multiple tasks in a block', () => {
    const input = `3/3/2026
Bug 12345: Fix login issue 1h
Task 67890: Update docs 30m

3/2/2026
Bug 111: Another bug 2h`;
    const tasks = parsePastedText(input);
    assert.equal(tasks.length, 3);
    assert.equal(tasks[0].date, '3/3/2026');
    assert.equal(tasks[0].number, '12345');
    assert.equal(tasks[2].date, '3/2/2026');
    assert.equal(tasks[2].number, '111');
  });

  it('parses mixed formats in one block', () => {
    const input = `3/3/2026
Task 27982: [UI/UX Redesign] Hover Over Tooltip
ask 31557: [Corporate Tags] Autocomplete
Bug 100: Fix crash 2h`;
    const tasks = parsePastedText(input);
    assert.equal(tasks.length, 3);
    assert.equal(tasks[0].type, 'Task');
    assert.equal(tasks[0].timeSpent, '');
    assert.equal(tasks[1].type, '');
    assert.equal(tasks[1].number, '31557');
    assert.equal(tasks[2].type, 'Bug');
    assert.equal(tasks[2].timeSpent, '2h');
  });

  it('preserves brackets and special chars in name', () => {
    const [t] = parsePastedText('3/3/2026\nTask 1: [Tag: Sub] Name (critical)');
    assert.equal(t.name, '[Tag: Sub] Name (critical)');
  });

  it('strips Pull Request prefix and saves to comments', () => {
    const [t] = parsePastedText('Pull Request 19082: Bug 31601: [Part Activity] Date Filter - the date pickers are disabled after selection of dates 1.5');
    assert.equal(t.type, 'Bug');
    assert.equal(t.number, '31601');
    assert.equal(t.name, '[Part Activity] Date Filter - the date pickers are disabled after selection of dates');
    assert.equal(t.timeSpent, '1.5h');
    assert.equal(t.comments, 'Pull Request 19082');
  });

  it('parses bare number as hours', () => {
    const [t] = parsePastedText('Bug 123: Fix login 2');
    assert.equal(t.timeSpent, '2h');
  });
});
