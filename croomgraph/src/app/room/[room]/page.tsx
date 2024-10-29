
"use client"
import { useRouter, useSearchParams } from 'next/navigation';
import axios from 'axios';
import Graph from '@/component/graphs/graphs';
import * as React from 'react';


const RoomPage = ({ params }: { params: { room: string } }) => {
    const router = useRouter();
    const { room } = params;
    const searchParams = useSearchParams();
    const indexPlaceStr = searchParams.get('index');
    const indexPlace = indexPlaceStr ? parseInt(indexPlaceStr, 0) : 0;
    const list: string[] = [];


    interface DataType {
        [key: string]: any;  // Define your actual data structure here
    }

    React.useEffect(() => {
        const feachData = () => {

            axios.get("http://192.168.23.61:8080/data").then(res => {
                console.log(res.data["data"][indexPlace]);
                console.log("HEllo");
                // const data: DataType = res.data["data"];
                // Object.entries(data).forEach(([key, value]) => {
                //     console.log(key,value)
                // })
            })
        }
        const interval = setInterval(feachData, 1000);
        return () => clearInterval(interval);
    }, [])

    return (<>

        <Graph />

    </>)
}

export default RoomPage;