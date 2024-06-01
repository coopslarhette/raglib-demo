import React, { useEffect, useState } from 'react'
import { codeToHtml } from 'shiki'

interface CodeBlockProps {
    language: string
    code: string
}

export default function CodeBlock({ language, code }: CodeBlockProps) {
    const [highlightedCode, setHighlightedCode] = useState('')

    useEffect(() => {
        const highlightCode = async () => {
            try {
                const highlighted = await codeToHtml(code, {
                    lang: language,
                    theme: 'vitesse-dark',
                })

                setHighlightedCode(highlighted)
            } catch (error) {
                console.error(`Language '${language}' not supported by Shiki`)
            }
        }

        highlightCode()
    }, [language, code])

    return <pre dangerouslySetInnerHTML={{ __html: highlightedCode }} />
}
