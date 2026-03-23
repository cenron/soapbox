interface NewPostsBannerProps {
  count: number
  onClick: () => void
}

export function NewPostsBanner({ count, onClick }: NewPostsBannerProps) {
  if (count === 0) return null

  const label = count === 1 ? "1 new post" : `${count} new posts`

  return (
    <button
      type="button"
      onClick={onClick}
      className="w-full border-b border-border bg-primary/5 px-4 py-3 text-center text-sm font-medium text-primary transition-colors hover:bg-primary/10"
    >
      Show {label}
    </button>
  )
}
