'use client'
import React from 'react'
import styles from './CitationBubble.module.css'

type CitationProps = {
    label: number
    onClick?: () => void
}

export function CitationBubble({ label, onClick }: CitationProps) {
    return (
        <span className={styles.citation} onClick={onClick}>
            {label}
        </span>
    )
}
