import { Component, type ReactNode } from 'react';
import mascot from '../assets/mascot.webp';

type Props = { children: ReactNode };
type State = { error: Error | null };

// ErrorBoundary catches render/lazy-load failures (e.g. a hashed chunk that
// 404s after a redeploy) so a single broken route shows a recoverable message
// instead of unmounting the whole app to a white screen.
export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  render() {
    if (this.state.error) {
      return (
        <div className="p-6">
          <div className="flex items-start gap-4 p-4 bg-red-500/10 border border-red-500/50 rounded-lg text-red-500 text-sm">
            {/* Decorative: the message beside it carries the meaning. */}
            <img src={mascot} alt="" width={72} height={75} className="w-18 shrink-0 hidden sm:block" />
            <div className="space-y-3">
              <p className="font-bold">Something went wrong loading this tool.</p>
              <p className="text-red-400 break-words">{this.state.error.message}</p>
              <button
                onClick={() => window.location.reload()}
                className="px-4 py-1.5 rounded bg-red-500/20 hover:bg-red-500/30 font-medium transition-colors"
              >
                Reload
              </button>
            </div>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}
