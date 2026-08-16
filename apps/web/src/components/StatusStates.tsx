export function LoadingLoom() {
  return (
    <div className="gantt-frame skeleton" aria-busy="true" aria-live="polite">
      <div className="banner">Threading the warp…</div>
      <div className="gantt-scroll">
        {Array.from({ length: 6 }).map((_, i) => (
          <div className="gantt-row" key={i}>
            <div className="line-meta">
              <code>——</code>
              <small>loading line</small>
            </div>
            <div className="warp" />
          </div>
        ))}
      </div>
    </div>
  );
}

export function EmptyLoom({ hint }: { hint: string }) {
  return (
    <div className="banner" role="status">
      <strong>No picks on this warp.</strong>
      <div className="muted">{hint}</div>
    </div>
  );
}

export function ErrorLoom({ message, onRetry }: { message: string; onRetry: () => void }) {
  return (
    <div className="banner error" role="alert">
      <strong>Loom dropped a thread.</strong>
      <div>{message}</div>
      <button className="btn btn-snag btn-sm mt-2" type="button" onClick={onRetry}>
        Restring
      </button>
    </div>
  );
}
