"use client"
import * as React from 'react';
import Box from '@mui/material/Box';
import Drawer from '@mui/material/Drawer';
import Button from '@mui/material/Button';
import List from '@mui/material/List';
import Divider from '@mui/material/Divider';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemIcon from '@mui/material/ListItemIcon';
import ListItemText from '@mui/material/ListItemText';
import axios from 'axios';


export default function TemporaryDrawer() {
    const [open, setOpen] = React.useState(false);
    const list: string[] = [];

    const toggleDrawer = (newOpen: boolean) => () => {
        setOpen(newOpen);
    };

    interface DataType {
        [key: string]: any;  // Define your actual data structure here
    }

    const [state, setState] = React.useState<string[]>([]);


    React.useEffect(() => {
        axios.get("http://192.168.23.61:8080/data").then(res => {
            const data: DataType = res.data["data"];
            Object.entries(data).forEach(([key, value]) => {
                list.push(value["Room"])
            })

            setState(list)


        })
    }, [])

    const RedirectToRoom: (roomName: string) => void = (roomName) => {
        console.log(`Redirecting to room: ${roomName}`);
      };


    const DrawerList = (
        <Box sx={{ width: 250 }} role="presentation" onClick={toggleDrawer(false)}>
            <List>
                {state.map((text, index) => {
                    console.log(text, "ffffff")
                    return (
                        <ListItem key={text} disablePadding>
                            <ListItemButton>
                                <ListItemText primary={text} onClick={()=>RedirectToRoom(text)}/>
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
            <Button onClick={toggleDrawer(true)}>Rooms</Button>
            <Drawer open={open} onClose={toggleDrawer(false)}>
                {DrawerList}
            </Drawer>
        </div>
    );
}


{/* <List> */ }
// {['All mail', 'Trash', 'Spam'].map((text, index) => (
//   <ListItem key={text} disablePadding>
// <ListItemButton>
//   {/* <ListItemIcon>
// {index % 2 === 0 ? <InboxIcon /> : <MailIcon />}
//   </ListItemIcon> */}
//   <ListItemText primary={text} />
// </ListItemButton>
//   </ListItem>
// ))}
// </List>