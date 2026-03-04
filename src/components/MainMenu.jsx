import React from 'react';
import { Box, Text, useApp, useInput } from 'ink';
import SelectInput from 'ink-select-input';

export const SHORTCUTS = {
  a: 'add',
  p: 'paste',
  s: 'summary',
  e: 'edit',
  q: 'exit',
};

const items = [
  { label: '(a) Add Task', value: 'add' },
  { label: '(p) Paste Tasks', value: 'paste' },
  { label: '(s) View Summary', value: 'summary' },
  { label: '(e) Edit / Delete', value: 'edit' },
  { label: '(q) Exit', value: 'exit' },
];

export default function MainMenu({ onSelect }) {
  const { exit } = useApp();

  const handleAction = (value) => {
    if (value === 'exit') {
      exit();
      return;
    }
    onSelect(value);
  };

  useInput((ch) => {
    const value = SHORTCUTS[ch];
    if (value) {
      handleAction(value);
    }
  });

  const handleSelect = (item) => {
    handleAction(item.value);
  };

  return (
    <Box flexDirection="column">
      <Text dimColor>Arrow keys + Enter or press shortcut key</Text>
      <Box marginTop={1}>
        <SelectInput items={items} onSelect={handleSelect} />
      </Box>
    </Box>
  );
}
