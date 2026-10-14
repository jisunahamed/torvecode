declare module "virtual:opencode-app-assets" {
  /** Reads the embedded web UI archive (see `app-archive.ts`); each build supplies it from its own embedded files. */
  const read: () => Uint8Array
  export default read
}

declare module "*.ansi" {
  const content: string
  export default content
}

declare module "*.txt" {
  const content: string
  export default content
}
