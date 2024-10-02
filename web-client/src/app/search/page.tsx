import React, { Suspense } from 'react'
import styles from './page.module.css'
import SearchContainer from './SearchContainer'
import { CircularProgress } from '@mui/material'

export default function SearchPage({
    searchParams,
}: {
    searchParams: { q?: string }
}) {
    return (
        <div className={styles.root}>
            <h2 className={styles.header}>RAGLib</h2>
            <Suspense
                fallback={
                    <CircularProgress style={{ color: 'var(--brand-teal)' }} />
                }
            >
                <SearchContainer initialQuery={searchParams.q || ''} />
            </Suspense>
        </div>
    )
}
