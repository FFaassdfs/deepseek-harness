import './style.css';
import {EventsOn} from '../wailsjs/runtime/runtime';
import {Retry} from '../wailsjs/go/main/App';

EventsOn('dsh-error', (msg) => {
    const status = document.getElementById('status');
    const spinner = document.getElementById('spinner');
    const hint = document.getElementById('hint');
    const retry = document.getElementById('retry');
    if (spinner) spinner.classList.add('hidden');
    if (status) {
        status.classList.add('error');
        status.textContent = '启动失败';
    }
    if (hint) {
        hint.textContent = msg;
    }
    if (retry) retry.classList.remove('hidden');
});

EventsOn('dsh-updating', (info) => {
    const status = document.getElementById('status');
    if (status) status.textContent = `正在更新 harness 到 ${info.latest}...`;
});

EventsOn('dsh-update', (info) => {
    const status = document.getElementById('status');
    if (status) status.textContent = `已更新到 ${info.latest}，正在启动...`;
});

document.getElementById('retry').addEventListener('click', () => {
    const spinner = document.getElementById('spinner');
    const status = document.getElementById('status');
    const hint = document.getElementById('hint');
    const retry = document.getElementById('retry');
    if (spinner) spinner.classList.remove('hidden');
    if (status) {
        status.classList.remove('error');
        status.textContent = '正在启动 DeepSeek Harness...';
    }
    if (hint) hint.textContent = '';
    if (retry) retry.classList.add('hidden');
    Retry();
});
