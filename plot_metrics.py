#!/usr/bin/env python3
"""
Plot CFR training metrics from CSV files.

Usage:
    python3 plot_metrics.py

Requires: matplotlib, pandas
    pip install matplotlib pandas
"""

import pandas as pd
import matplotlib.pyplot as plt
import glob
import os

DATA_DIR = 'data'

def plot_exploitability():
    """Plot exploitability over iterations."""
    filepath = os.path.join(DATA_DIR, 'exploitability.csv')
    if not os.path.exists(filepath):
        print(f"{filepath} not found, skipping...")
        return

    df = pd.read_csv(filepath)

    plt.figure(figsize=(10, 6))
    plt.plot(df['iteration'], df['exploitability'], 'b-', linewidth=2)
    plt.xlabel('Iteration', fontsize=12)
    plt.ylabel('Exploitability (BB)', fontsize=12)
    plt.title('Strategy Exploitability Over Training', fontsize=14)
    plt.grid(True, alpha=0.3)
    plt.tight_layout()
    plt.savefig(os.path.join(DATA_DIR, 'exploitability.png'), dpi=150)
    plt.close()
    print(f"Saved: {DATA_DIR}/exploitability.png")

def plot_bet_frequency():
    """Plot OOP (P1) QQ first-action bet frequency: current vs average strategy."""
    filepath = os.path.join(DATA_DIR, 'bet_frequency.csv')
    if not os.path.exists(filepath):
        print(f"{filepath} not found, skipping...")
        return

    df = pd.read_csv(filepath)

    # Get columns
    current_cols = [c for c in df.columns if c.startswith('current_')]
    avg_cols = [c for c in df.columns if c.startswith('average_')]

    if not current_cols and not avg_cols:
        print("No bet frequency data to plot")
        return

    # Create single plot with current vs average
    plt.figure(figsize=(12, 6))

    # Plot current strategy (lighter, shows variance)
    for col in current_cols:
        label = col.replace('current_bet_freq_', 'Current ')
        plt.plot(df['iteration'], df[col], alpha=0.5, linewidth=1, label=label)

    # Plot average strategy (darker, shows convergence)
    for col in avg_cols:
        label = col.replace('average_bet_freq_', 'Average ')
        plt.plot(df['iteration'], df[col], linewidth=2.5, label=label)

    plt.xlabel('Iteration', fontsize=12)
    plt.ylabel('Bet Frequency', fontsize=12)
    plt.title('OOP (P1) QQ First-Action Bet Frequency\nCurrent Strategy vs Average Strategy', fontsize=14)
    plt.legend(fontsize=9, loc='best')
    plt.grid(True, alpha=0.3)
    plt.ylim(0, 1)
    plt.tight_layout()
    plt.savefig(os.path.join(DATA_DIR, 'qq_bet_frequency.png'), dpi=150)
    plt.close()
    print(f"Saved: {DATA_DIR}/qq_bet_frequency.png")

def plot_combined():
    """Create combined 2-panel plot: exploitability + bet frequency."""
    fig, axes = plt.subplots(1, 2, figsize=(14, 5))

    # Panel 1: Exploitability
    exploitability_path = os.path.join(DATA_DIR, 'exploitability.csv')
    if os.path.exists(exploitability_path):
        df = pd.read_csv(exploitability_path)
        ax1 = axes[0]
        ax1.plot(df['iteration'], df['exploitability'], 'b-', linewidth=2)
        ax1.set_xlabel('Iteration', fontsize=11)
        ax1.set_ylabel('Exploitability (BB)', fontsize=11)
        ax1.set_title('Strategy Exploitability', fontsize=12)
        ax1.grid(True, alpha=0.3)

    # Panel 2: Bet Frequency
    bet_freq_path = os.path.join(DATA_DIR, 'bet_frequency.csv')
    if os.path.exists(bet_freq_path):
        df = pd.read_csv(bet_freq_path)
        ax2 = axes[1]

        current_cols = [c for c in df.columns if c.startswith('current_')]
        avg_cols = [c for c in df.columns if c.startswith('average_')]

        # Average all QQ combos together for cleaner view
        if current_cols:
            current_avg = df[current_cols].mean(axis=1)
            ax2.plot(df['iteration'], current_avg, 'r-', alpha=0.5, linewidth=1.5, label='Current Strategy')

        if avg_cols:
            avg_avg = df[avg_cols].mean(axis=1)
            ax2.plot(df['iteration'], avg_avg, 'b-', linewidth=2.5, label='Average Strategy')

        ax2.set_xlabel('Iteration', fontsize=11)
        ax2.set_ylabel('Bet Frequency', fontsize=11)
        ax2.set_title('OOP QQ First-Action Bet Frequency', fontsize=12)
        ax2.legend(fontsize=10)
        ax2.grid(True, alpha=0.3)
        ax2.set_ylim(0, 1)

    plt.tight_layout()
    plt.savefig(os.path.join(DATA_DIR, 'training_summary.png'), dpi=150)
    plt.close()
    print(f"Saved: {DATA_DIR}/training_summary.png")

def main():
    print("Generating CFR training plots...")
    print("-" * 40)

    if not os.path.exists(DATA_DIR):
        print(f"Error: {DATA_DIR}/ directory not found. Run the solver first.")
        return

    plot_exploitability()
    plot_bet_frequency()
    plot_combined()

    print("-" * 40)
    print(f"Done! Generated PNG files in {DATA_DIR}/:")
    print(f"  - {DATA_DIR}/exploitability.png     (exploitability convergence)")
    print(f"  - {DATA_DIR}/qq_bet_frequency.png   (current vs average strategy)")
    print(f"  - {DATA_DIR}/training_summary.png   (combined 2-panel view)")

if __name__ == '__main__':
    main()
