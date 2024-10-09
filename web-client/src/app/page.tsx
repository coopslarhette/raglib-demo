'use client'

import React from 'react'
import { useRouter } from 'next/navigation'
import styles from './page.module.css'
import SearchBar from '@/app/search/SearchBar'

export default function SearchHome() {
    const router = useRouter()

    const handleSubmit = (q: string) => {
        if (q) {
            router.push(`/search?q=${encodeURIComponent(q)}`)
        }
    }

    return (
        <div className={styles.container}>
            <h1 className={styles.title}>RAGLib Search</h1>
            <SearchBar onSearch={handleSubmit} />
        </div>
    )
}
