import React, { useEffect } from 'react'
import Prism from 'prismjs'
import 'prismjs/themes/prism.css' // or any other theme you prefer

interface CodeBlockProps {
    language: string
    code: string
}

export default function CodeBlock({ language, code }: CodeBlockProps) {
    useEffect(() => {
        Prism.highlightAll()
    }, [])

    return (
        <pre>
            <code className={`language-${language}`}>{code}</code>
        </pre>
    )
}
