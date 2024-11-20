'use client'

import React from 'react'
import { useRouter } from 'next/navigation'
import styles from './page.module.css'
import SearchBar from '@/app/search/SearchBar'
import SuggestedQueries from '@/app/search/SuggestedQueries'

export default function SearchHome() {
    const router = useRouter()

    const handleSubmit = (q: string) => {
        if (q) {
            router.push(`/search?q=${encodeURIComponent(q)}`)
        }
    }

    return (
        <div className={styles.container}>
            <h1 className={styles.title}>RAGLib Research</h1>
            <SearchBar onSearch={handleSubmit} />
            <SuggestedQueries onQueryClick={handleSubmit} />
        </div>
    )
}
