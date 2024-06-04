import React, { useEffect, useState } from 'react'
import { codeToHtml } from 'shiki'
import styles from './code-block.module.css'

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
                    theme: 'material-theme',
                })

                setHighlightedCode(highlighted)
            } catch (error) {
                console.error(`Language '${language}' not supported by Shiki`)
            }
        }

        highlightCode()
    }, [language, code])

    return <div className={styles.code} dangerouslySetInnerHTML={{ __html: highlightedCode }} />
}
