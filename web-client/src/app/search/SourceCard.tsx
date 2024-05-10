import React from 'react'
import styles from './SourceCard.module.css'
import {
    Card,
    CardActionArea,
    CardContent,
    CardMedia,
    Link,
    Typography,
} from '@mui/material'
import { WebReference } from '@/app/search/types'

type SourceCardProps = {
    source: WebReference
}

export function SourceCard({
    source: { title, link, displayedLink, date, author, favicon, thumbnail },
}: SourceCardProps) {
    const handleClick = () => {
        window.open(link, '_blank', 'noopener noreferrer')
    }

    return (
        <Card>
            <CardActionArea onClick={handleClick} className={styles.card}>
                {thumbnail && (
                    <CardMedia
                        component="img"
                        height="140"
                        image={thumbnail}
                        alt="Thumbnail"
                    />
                )}
                <CardContent className={styles.cardContent}>
                    <div className={styles.header}>
                        {favicon && (
                            <img
                                src={favicon}
                                alt="Favicon"
                                className={styles.favicon}
                            />
                        )}
                        <Typography
                            className={styles.displayedLink}
                            variant="body2"
                            color="text.secondary"
                        >
                            {displayedLink}
                        </Typography>
                    </div>
                    <Typography variant="subtitle1" component="div">
                        {title}
                    </Typography>
                    <div className={styles.metadata}>
                        {author && (
                            <Typography variant="body2">
                                Author: {author}
                            </Typography>
                        )}
                        {date && (
                            <Typography variant="body2">{date}</Typography>
                        )}
                    </div>
                </CardContent>
            </CardActionArea>
        </Card>
    )
}
