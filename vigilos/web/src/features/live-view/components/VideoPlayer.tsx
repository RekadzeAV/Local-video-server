export type VideoPlayerProps = { src: string };

// Placeholder non-UI implementation to satisfy lints without React deps.
export function VideoPlayer(props: VideoPlayerProps): string {
  return `Video placeholder for ${props.src}`;
}

