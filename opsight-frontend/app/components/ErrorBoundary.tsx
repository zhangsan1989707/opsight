'use client';

import React, { useState, useEffect, Component, ReactNode } from 'react';

interface ErrorBoundaryProps {
  children: ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="bg-surface-50 border border-red-500/20 rounded-xl p-6 my-4">
          <div className="flex items-start gap-3">
            <div className="w-8 h-8 rounded-full bg-red-500/10 flex items-center justify-center flex-shrink-0">
              <svg width="16" height="16" fill="none" stroke="#ef4444" strokeWidth="2" viewBox="0 0 16 16">
                <circle cx="8" cy="8" r="6" />
                <path d="M8 5v3M8 10v1" />
              </svg>
            </div>
            <div>
              <h3 className="text-sm font-medium text-red-400">Something went wrong</h3>
              <p className="text-xs text-zinc-500 mt-1 font-mono">{this.state.error?.message}</p>
              <button
                onClick={() => this.setState({ hasError: false, error: null })}
                className="mt-3 text-xs text-accent hover:text-accent/80 transition-colors"
              >
                Try again
              </button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export { ErrorBoundary };
