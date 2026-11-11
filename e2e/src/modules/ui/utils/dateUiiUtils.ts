/**
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

function formatDuration(seconds: number): string {
    const days = Math.floor(seconds / (24 * 60 * 60));
    seconds %= 24 * 60 * 60;
    const hours = Math.floor(seconds / (60 * 60));
    seconds %= 60 * 60;
    const minutes = Math.floor(seconds / 60);
    let result = '';

    if (days > 0) result += `${days}d `;
    if (hours > 0) result += `${hours}h `;
    if (minutes > 0) result += `${minutes}min`;
    if (result === '') result = '0min';
    return result.trim();
}

function convertSecondsToTimeUnits(seconds: number): { days: number; hours: number; minutes: number } {
    const days = Math.floor(seconds / (24 * 60 * 60));
    const hours = Math.floor((seconds % (24 * 60 * 60)) / (60 * 60));
    const minutes = Math.floor(((seconds % (24 * 60 * 60)) % (60 * 60)) / 60);

    return { days, hours, minutes };
}

function hoursToSeconds(hours: number): number {
    return hours * 60 * 60;
}

function convertTimeUnitsToSeconds(
    days: number,
    hours: number,
    minutes: number
): { daysInSeconds: number; hoursInSeconds: number; minutesInSeconds: number } {
    const daysInSeconds = days * 24 * 60 * 60;
    const hoursInSeconds = hours * 60 * 60;
    const minutesInSeconds = minutes * 60;
    return { daysInSeconds, hoursInSeconds, minutesInSeconds };
}

export const dateUiUtils = {
    formatDuration,
    convertSecondsToTimeUnits,
    convertTimeUnitsToSeconds,
    hoursToSeconds
};
