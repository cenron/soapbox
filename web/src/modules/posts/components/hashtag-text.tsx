import { Link } from "react-router"

interface HashtagTextProps {
  body: string
}

export function HashtagText({ body }: HashtagTextProps) {
  const parts = body.split(/(#\w+)/g)

  return (
    <>
      {parts.map((part, i) => {
        if (part.startsWith("#")) {
          const tag = part.slice(1)
          return (
            <Link
              key={i}
              to={`/search?q=%23${tag}`}
              className="text-primary hover:underline"
              onClick={(e) => e.stopPropagation()}
            >
              {part}
            </Link>
          )
        }
        return <span key={i}>{part}</span>
      })}
    </>
  )
}
