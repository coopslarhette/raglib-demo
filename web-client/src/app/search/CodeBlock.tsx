import React from 'react'
import styles from './code-block.module.css'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'

const arcTheme: { [key: string]: React.CSSProperties } = {
    'code[class*="language-"]': {
        color: '#E4E4E8',
        backgroundColor: '#1C1C1E',
        fontFamily: 'ui-monospace, SFMono, Menlo, monospace',
        fontSize: '14px',
        textAlign: 'left',
        whiteSpace: 'pre',
        wordSpacing: 'normal',
        wordBreak: 'normal',
        wordWrap: 'normal',
        lineHeight: 1.6,
        tabSize: 2,
        margin: 0,
    },
    // Comments
    comment: {
        color: '#757575',
        fontStyle: 'italic',
    },
    prolog: {
        color: '#757575',
    },
    doctype: {
        color: '#757575',
    },
    cdata: {
        color: '#757575',
    },

    // String-related
    string: {
        color: '#FF9580',
    },
    'attr-value': {
        color: '#FF9580',
    },
    char: {
        color: '#FF9580',
    },
    'template-string': {
        color: '#FF9580',
    },

    // Keywords and operators
    keyword: {
        color: '#82AAFF',
    },
    builtin: {
        color: '#82AAFF',
    },
    'class-name': {
        color: '#FFB86C',
    },
    function: {
        color: '#77E6B0',
    },
    method: {
        color: '#77E6B0',
    },
    operator: {
        color: '#E4E4E8',
    },

    // Language-specific tokens
    constant: {
        color: '#FF757F',
    },
    symbol: {
        color: '#FF757F',
    },
    regex: {
        color: '#F2C4CE',
    },
    'attr-name': {
        color: '#77E6B0',
    },

    // Variables and properties
    variable: {
        color: '#E4E4E8',
    },
    property: {
        color: '#77E6B0',
    },
    parameter: {
        color: '#E4E4E8',
    },

    // Numbers and types
    number: {
        color: '#FFB86C',
    },
    boolean: {
        color: '#FF757F',
    },
    tag: {
        color: '#FF757F',
    },

    // Punctuation
    punctuation: {
        color: '#858585',
    },

    // Go-specific
    namespace: {
        color: '#82AAFF',
    },
    package: {
        color: '#82AAFF',
    },
    type: {
        color: '#FFB86C',
    },
    interface: {
        color: '#FFB86C',
    },

    // Markup
    important: {
        color: '#FF757F',
        fontWeight: 'bold',
    },
    bold: {
        fontWeight: 'bold',
    },
    italic: {
        fontStyle: 'italic',
    },
}

interface CodeBlockProps {
    language?: string | undefined
    code: string
}

export default function CodeBlock({ language, code }: CodeBlockProps) {
    return (
        <div className={styles.wrapper}>
            <SyntaxHighlighter
                PreTag="div"
                children={String(code).replace(/\n$/, '')}
                language={language}
                style={arcTheme}
                customStyle={{
                    margin: 0,
                    padding: '1rem',
                    borderRadius: '8px',
                    backgroundColor: '#1C1C1E',
                }}
            />
        </div>
    )
}
