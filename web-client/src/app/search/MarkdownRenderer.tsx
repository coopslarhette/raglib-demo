import ReactMarkdown, { Components as MarkdownComponents } from 'react-markdown'
import remarkGfm from 'remark-gfm'
import rehypeRaw from 'rehype-raw'
import { CitationBubble } from '@/app/search/CitationBubble'
import CodeBlock from '@/app/search/CodeBlock'
import React from 'react'
import styles from './MarkdownRenderer.module.css'

interface CitedComponent {
    cited: React.ComponentType<{
        children: number
        className?: string
    }>
}

type Components = Partial<MarkdownComponents & CitedComponent>

interface MarkdownRendererProps {
    content: string
    setHoveredCitationIndex: React.Dispatch<React.SetStateAction<number | null>>
}

const baseComponents: Components = {
    code: ({ className, children }) => {
        const language = className?.match(/language-(\w+)/)?.[1]

        if (!children) {
            return null
        }

        if (!language) {
            return <code className={styles.mdInlineCode}>{children}</code>
        }

        return <CodeBlock code={String(children)} language={language} />
    },
    // Use these custom components to apply styles correctly
    ol: ({ children, ...props }) => (
        <ol className={styles.mdList} {...props}>
            {children}
        </ol>
    ),
    ul: ({ children, ...props }) => (
        <ul className={styles.mdList} {...props}>
            {children}
        </ul>
    ),
    p: ({ children, ...props }) => (
        <p className={styles.block} {...props}>
            {children}
        </p>
    ),
    blockquote: ({ children, ...props }) => (
        <blockquote className={styles.block} {...props}>
            {children}
        </blockquote>
    ),
    dl: ({ children, ...props }) => (
        <dl className={styles.block} {...props}>
            {children}
        </dl>
    ),
    table: ({ children, ...props }) => (
        <table className={styles.block} {...props}>
            {children}
        </table>
    ),
    pre: ({ children, ...props }) => (
        <pre className={styles.block} {...props}>
            {children}
        </pre>
    ),
    details: ({ children, ...props }) => (
        <details className={styles.block} {...props}>
            {children}
        </details>
    ),
}

export function MarkdownRenderer({
    content,
    setHoveredCitationIndex,
}: MarkdownRendererProps) {
    const componentsToUse: Components = {
        ...baseComponents,
        cited: ({ children }) => {
            if (!children || isNaN(Number(children))) return null

            return (
                <CitationBubble
                    citationNumber={Number(children)}
                    onMouseLeave={() => setHoveredCitationIndex(null)}
                    onMouseEnter={() =>
                        setHoveredCitationIndex(Number(children))
                    }
                />
            )
        },
    }

    return (
        <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            rehypePlugins={[rehypeRaw]}
            components={componentsToUse}
            className={styles.markdownContainer}
        >
            {content}
        </ReactMarkdown>
    )
}
