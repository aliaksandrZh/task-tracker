export const DATE_PATTERNS = [
  /^(\d{1,2}\/\d{1,2}\/\d{4})$/,
  /^(\d{4}-\d{2}-\d{2})$/,
];

export const PREFIX_PATTERN = /^Pull\s+Request\s+\d+\s*:\s*/i;

export const TYPE_PATTERNS = [
  { re: /^(Bug|Task)\b/i, extract: (m) => m[1] },
];

export const NUMBER_PATTERNS = [
  { re: /^#?(\d+)\s*:?\s*/, extract: (m) => m[1] },
];

export const TIME_PATTERNS = [
  { re: /\s+(\d+(?:\.\d+)?h\s*\d+(?:\.\d+)?m)$/, extract: (m) => m[1] },
  { re: /\s+(\d+(?:\.\d+)?[hm])$/, extract: (m) => m[1] },
  { re: /\s+(\d+(?:\.\d+)?)$/, extract: (m) => `${m[1]}h` },
];

function normalizeType(type) {
  return type.charAt(0).toUpperCase() + type.slice(1).toLowerCase();
}

function todayStr() {
  const d = new Date();
  return `${d.getMonth() + 1}/${d.getDate()}/${d.getFullYear()}`;
}

export function tryPatterns(patterns, text) {
  for (const p of patterns) {
    const m = text.match(p.re || p);
    if (m) return { match: m, extract: p.extract ? p.extract(m) : m[1] };
  }
  return null;
}

export function parseLine(line) {
  let remaining = line;
  let type = '';
  let number = '';
  let timeSpent = '';
  let comments = '';

  // Strip "Pull Request XXXXX:" prefix, save to comments
  const prMatch = remaining.match(PREFIX_PATTERN);
  if (prMatch) {
    comments = prMatch[0].trim().replace(/:\s*$/, '');
    remaining = remaining.slice(prMatch[0].length).trim();
  }

  const typeResult = tryPatterns(TYPE_PATTERNS, remaining);
  if (typeResult) {
    type = normalizeType(typeResult.extract);
    remaining = remaining.slice(typeResult.match[0].length).trim();
  } else {
    const unknownWord = remaining.match(/^\S+\s+(?=\d)/);
    if (unknownWord) {
      remaining = remaining.slice(unknownWord[0].length).trim();
    }
  }

  const numberResult = tryPatterns(NUMBER_PATTERNS, remaining);
  if (numberResult) {
    number = numberResult.extract;
    remaining = remaining.slice(numberResult.match[0].length).trim();
  }

  const timeResult = tryPatterns(TIME_PATTERNS, remaining);
  if (timeResult) {
    timeSpent = timeResult.extract;
    remaining = remaining.slice(0, -timeResult.match[0].length).trim();
  }

  const name = remaining;

  return { type, number, name, timeSpent, comments };
}

export function parsePastedText(text) {
  const lines = text.split('\n').map(l => l.trim()).filter(Boolean);
  const tasks = [];
  let currentDate = '';

  for (const line of lines) {
    const dateResult = tryPatterns(DATE_PATTERNS.map(re => ({ re, extract: (m) => m[1] })), line);
    if (dateResult) {
      currentDate = dateResult.extract;
      continue;
    }

    const { type, number, name, timeSpent, comments } = parseLine(line);

    const missing = [];
    if (!currentDate) missing.push('date');
    if (!type) missing.push('type');
    if (!number) missing.push('number');
    if (!name) missing.push('name');
    if (!timeSpent) missing.push('timeSpent');

    tasks.push({
      date: currentDate || todayStr(),
      type,
      number,
      name,
      timeSpent,
      comments,
      missing,
    });
  }

  return tasks;
}
