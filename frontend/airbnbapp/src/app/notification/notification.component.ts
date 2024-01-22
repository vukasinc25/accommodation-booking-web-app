import { Component, OnInit } from '@angular/core';
import { NotificationService } from '../service/notification.service';
import { AuthService } from '../service/auth.service';
import { Notification1 } from '../model/notification';

@Component({
  selector: 'app-notification',
  templateUrl: './notification.component.html',
  styleUrls: ['./notification.component.css']
})
export class NotificationComponent implements OnInit{
  notifications: Notification1[] = [];

  constructor(
    private notificationService: NotificationService,
    private authService: AuthService
    ) {}

  ngOnInit(): void {
    this.loadNotifications();
  }

  loadNotifications(): void {
    this.notificationService.getAllByHostId(this.authService.getUserId()).subscribe(
      (data) => {
        console.log(data)
        this.notifications = data
      },
      (error) => {
        console.error('Error getting notifications')
      }
    )
  }
}
