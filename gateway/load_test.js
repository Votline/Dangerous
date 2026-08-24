import exec from 'k6/x/exec';
import { check } from 'k6';

export const options = {
  scenarios: {
    users_test: {
      executor: 'constant-vus',
      exec: 'runUsersCheck',
      vus: 5,
      duration: '30s',
    },
    marks_test: {
      executor: 'constant-vus',
      exec: 'runMarksCheck',
      vus: 5,
      duration: '30s',
    },
  },
};

export function runUsersCheck() {
  let success = false;
  try {
    exec.command('gcli', ['run', 'users_check.gurlf', '-dw', '-dp']);
    success = true;
  } catch (e) {
    success = false;
  }

  check(null, {
    'users_check exit code 0': () => success,
  });
}

export function runMarksCheck() {
  let success = false;
  try {
    exec.command('gcli', ['run', 'marks_check.gurlf', '-dw', '-dp']);
    success = true;
  } catch (e) {
    success = false;
  }

  check(null, {
    'marks_check exit code 0': () => success,
  });
}
