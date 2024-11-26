"use client"
import * as React from 'react';
import Box from '@mui/material/Box';
import Drawer from '@mui/material/Drawer';
// import Button from '@mui/material/Button';
import List from '@mui/material/List';
import Divider from '@mui/material/Divider';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import { useRouter } from 'next/navigation';
import axios from 'axios';
import Image from 'next/image'

import RoomsPng from "../../public/images/room.png"




export default function TemporaryDrawer() {
    const apiUrl = process.env.NEXT_PUBLIC_API_URL;
    const [open, setOpen] = React.useState(false);
    // const list: string[] = [];
    const router = useRouter();
    const toggleDrawer = (newOpen: boolean) => () => {
        setOpen(newOpen);
    };

    // interface DataType {
    //     [key: string]: any;  // Define your actual data structure here
    // }

    const [state, setState] = React.useState<string[]>([]);


    React.useEffect(() => {
        console.log(apiUrl+"/getrooms")
        axios.get(apiUrl+"/getrooms").then(res => {
            setState(res.data["data"])
        })
    }, [])

    const RedirectToRoom: (roomName: string, indexPlace: number) => void = (roomName, indexPlace) => {
        router.push(`/room/${roomName}?index=${indexPlace}`);
    };


    const DrawerList = (
        <Box sx={{ width: 250 }} role="presentation" onClick={toggleDrawer(false)}>
            <List>
                {state.map((text, index) => {
                   
                    return (
                        <ListItem key={text} disablePadding>
                            <ListItemButton>
                                <ListItemText primary={text} onClick={() => RedirectToRoom(text, index)} />
                            </ListItemButton>
                        </ListItem>
                    );
                })}

            </List>
            <Divider />
        </Box>
    );

    return (
        <div>
            {/* <Button onClick={toggleDrawer(true)}>Rooms</Button> */}
            {/* <img onClick={toggleDrawer(true)}  */}
             <Image onClick={toggleDrawer(true)} src={RoomsPng} alt="Room" width="50" height="50"/> 
            <Drawer open={open} onClose={toggleDrawer(false)}>
                {DrawerList}
            </Drawer>
        </div>
    );
}


