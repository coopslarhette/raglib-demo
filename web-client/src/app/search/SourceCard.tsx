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
            <CardActionArea onClick={handleClick}>
                {thumbnail && (
                    <CardMedia
                        component="img"
                        height="140"
                        image={thumbnail}
                        alt="Thumbnail"
                    />
                )}
                <CardContent>
                    <div className={styles.header}>
                        {favicon && (
                            <img
                                src={favicon}
                                alt="Favicon"
                                className={styles.favicon}
                            />
                        )}
                        <Typography variant="h6" component="div">
                            {title}
                        </Typography>
                    </div>
                    <Typography variant="body2" color="text.secondary">
                        {displayedLink}
                    </Typography>
                    <div className={styles.metadata}>
                        {author && (
                            <Typography variant="body2">
                                Author: {author}
                            </Typography>
                        )}
                        {date && (
                            <Typography variant="body2">
                                Date: {date}
                            </Typography>
                        )}
                    </div>
                </CardContent>
            </CardActionArea>
        </Card>
    )
}
