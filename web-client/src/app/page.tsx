'use client'

import React, { useState } from 'react'
import { useRouter } from 'next/navigation'
import styles from './page.module.css'
import SearchBar from '@/app/search/SearchBar'

export default function SearchHome() {
    const [query, setQuery] = useState('')
    const router = useRouter()

    const handleSubmit = (e?: React.FormEvent) => {
        e?.preventDefault()
        const q = query.trim()

        if (q) {
            router.push(`/search?q=${encodeURIComponent(q)}`)
        }
    }

    return (
        <div className={styles.container}>
            <h1 className={styles.title}>RAGLib Search</h1>
            <SearchBar
                query={query}
                setQuery={setQuery}
                onSearch={handleSubmit}
            />
        </div>
    )
}
